package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	adopipelines "github.com/kyma-project/test-infra/pkg/azuredevops/pipelines"
	"github.com/kyma-project/test-infra/pkg/sets"
	"github.com/kyma-project/test-infra/pkg/sign"
	"github.com/kyma-project/test-infra/pkg/tags"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/pipelines"
	"golang.org/x/net/context"
	errutil "k8s.io/apimachinery/pkg/util/errors"
)

type options struct {
	Config     `json:"config"`
	ConfigPath string       `json:"config_path"`
	Context    string       `json:"context"`
	Dockerfile string       `json:"dockerfile"`
	EnvFile    string       `json:"env_file"`
	Name       string       `json:"name"`
	Variant    string       `json:"variant"`
	LogDir     string       `json:"log_dir"`
	OrgRepo    string       `json:"org_repo"`
	Silent     bool         `json:"Silent"`
	IsCI       bool         `json:"is_ci"`
	Tags       sets.Tags    `json:"tags"`
	BuildArgs  sets.Tags    `json:"build_args"`
	Platforms  sets.Strings `json:"platforms"`
	ExportTags bool         `json:"export_tags"`
	// SignOnly enables only signing of images. No build will be performed.
	SignOnly      bool         `json:"sign_only"`
	ImagesToSign  imagesToSign `json:"images_to_sign"`
	BuildInADO    bool         `json:"build_in_ado"`
	ParseTagsOnly bool         `json:"parse_tags_only"`
	Debug         bool         `json:"debug"`
}

const (
	PlatformLinuxAmd64 = "linux/amd64"
	PlatformLinuxArm64 = "linux/arm64"
)

type imagesToSign []string

func (i *imagesToSign) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *imagesToSign) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// parseVariable returns a build-arg.
// Keys are set to upper-case.
func parseVariable(key, value string) string {
	k := strings.TrimSpace(key)
	return k + "=" + strings.TrimSpace(value)
}

// runInBuildKit prepares command execution and handles gathering logs from BuildKit-enabled run
// This function is used only in customized environment
func runInBuildKit(o options, name string, destinations, platforms []string, buildArgs map[string]string) error {
	dockerfile := filepath.Base(o.Dockerfile)
	dockerfileDir := filepath.Dir(o.Dockerfile)
	args := []string{
		"build", "--frontend=dockerfile.v0",
		"--local", "context=" + o.Context,
		"--local", "dockerfile=" + filepath.Join(o.Context, dockerfileDir),
		"--opt", "filename=" + dockerfile,
	}

	// output definition, multiple images support
	args = append(args, "--output", "type=image,\"name="+strings.Join(destinations, ",")+"\",push=true")

	// build-args
	for k, v := range buildArgs {
		args = append(args, "--opt", "build-arg:"+parseVariable(k, v))
	}

	if len(platforms) > 0 {
		args = append(args, "--opt", "platform="+strings.Join(platforms, ","))
	}

	if o.Cache.Enabled {
		// TODO (@Ressetkk): Implement multiple caches, see https://github.com/moby/buildkit#export-cache
		args = append(args,
			"--export-cache", "type=registry,ref="+o.Cache.CacheRepo,
			"--import-cache", "type=registry,ref="+o.Cache.CacheRepo)
	}

	cmd := exec.Command("buildctl-daemonless.sh", args...)

	var outw []io.Writer
	var errw []io.Writer

	if !o.Silent {
		outw = append(outw, os.Stdout)
		errw = append(errw, os.Stderr)
	}

	f, err := os.Create(filepath.Join(o.LogDir, strings.TrimSpace("build_"+strings.TrimSpace(name)+".log")))
	if err != nil {
		return fmt.Errorf("could not create log file: %w", err)
	}

	outw = append(outw, f)
	errw = append(errw, f)

	cmd.Stdout = io.MultiWriter(outw...)
	cmd.Stderr = io.MultiWriter(errw...)

	return cmd.Run()
}

// runInKaniko prepares command execution and handles gathering logs to file
func runInKaniko(o options, name string, destinations, platforms []string, buildArgs map[string]string) error {
	args := []string{
		"--context=" + o.Context,
		"--dockerfile=" + o.Dockerfile,
	}
	for _, dst := range destinations {
		args = append(args, "--destination="+dst)
	}

	for k, v := range buildArgs {
		args = append(args, "--build-arg="+parseVariable(k, v))
	}

	if len(platforms) > 0 {
		fmt.Println("'--platform' parameter not supported in kaniko-mode. Use buildkit-enabled image")
	}

	if o.Config.Cache.Enabled {
		args = append(args, "--cache="+strconv.FormatBool(o.Cache.Enabled),
			"--cache-copy-layers="+strconv.FormatBool(o.Cache.CacheCopyLayers),
			"--cache-run-layers="+strconv.FormatBool(o.Cache.CacheRunLayers),
			"--cache-repo="+o.Cache.CacheRepo)
	}

	if o.Config.LogFormat != "" {
		args = append(args, "--log-format="+o.Config.LogFormat)
	}

	if o.Config.Reproducible {
		args = append(args, "--reproducible=true")
	}

	cmd := exec.Command("/kaniko/executor", args...)

	var outw []io.Writer
	var errw []io.Writer

	if !o.Silent {
		outw = append(outw, os.Stdout)
		errw = append(errw, os.Stderr)
	}

	f, err := os.Create(filepath.Join(o.LogDir, strings.TrimSpace("build_"+strings.TrimSpace(name)+".log")))
	if err != nil {
		return fmt.Errorf("could not create log file: %w", err)
	}

	outw = append(outw, f)
	errw = append(errw, f)

	cmd.Stdout = io.MultiWriter(outw...)
	cmd.Stderr = io.MultiWriter(errw...)

	return cmd.Run()
}

func prepareADOTemplateParameters(options options) (adopipelines.OCIImageBuilderTemplateParams, error) {
	templateParameters := make(adopipelines.OCIImageBuilderTemplateParams)

	repo, present := os.LookupEnv("REPO_NAME")
	if !present {
		return nil, fmt.Errorf("REPO_NAME environment variable is not set, please set it to valid repository name")
	}
	templateParameters.SetRepoName(repo)

	owner, present := os.LookupEnv("REPO_OWNER")
	if !present {
		return nil, fmt.Errorf("REPO_OWNER environment variable is not set, please set it to valid repository owner")
	}
	templateParameters.SetRepoOwner(owner)

	jobType, present := os.LookupEnv("JOB_TYPE")
	if !present {
		return nil, fmt.Errorf("JOB_TYPE environment variable is not set, please set it to valid job type")
	}
	if jobType == "presubmit" {
		templateParameters.SetPresubmitJobType()
	} else if jobType == "postsubmit" {
		templateParameters.SetPostsubmitJobType()
	} else {
		return nil, fmt.Errorf("JOB_TYPE environment variable is not set to valid value, please set it to either 'presubmit' or 'postsubmit'")
	}

	number, present := os.LookupEnv("PULL_NUMBER")
	if !present {
		return nil, fmt.Errorf("PULL_NUMBER environment variable is not set, please set it to valid pull request number")
	}
	templateParameters.SetPullNumber(number)

	baseSHA, present := os.LookupEnv("PULL_BASE_SHA")
	if !present {
		return nil, fmt.Errorf("PULL_BASE_SHA environment variable is not set, please set it to valid pull base SHA")
	}
	templateParameters.SetBaseSHA(baseSHA)

	pullSHA, present := os.LookupEnv("PULL_PULL_SHA")
	if present {
		templateParameters.SetPullSHA(pullSHA)
	}

	templateParameters.SetImageName(options.Name)

	templateParameters.SetDockerfilePath(options.Dockerfile)

	templateParameters.SetBuildContext(options.Context)

	templateParameters.SetExportTags(options.ExportTags)

	if len(options.BuildArgs) > 0 {
		templateParameters.SetBuildArgs(options.BuildArgs.String())
	}

	if len(options.Tags) > 0 {
		templateParameters.SetImageTags(options.Tags.String())
	}

	err := templateParameters.Validate()
	if err != nil {
		return nil, fmt.Errorf("failed validating ADO template parameters, err: %w", err)
	}

	return templateParameters, nil
}

// TODO(dekiel): refactor this function to accept clients as parameters to make it testable with mocks.
func buildInADO(o options) error {
	fmt.Println("Building image in ADO pipeline.")
	adoPAT, present := os.LookupEnv("ADO_PAT")
	if !present {
		return fmt.Errorf("build in ADO failed, ADO_PAT environment variable is not set, please set it to valid ADO PAT")
	}

	fmt.Println("Preparing ADO template parameters.")
	templateParameters, err := prepareADOTemplateParameters(o)
	if err != nil {
		return fmt.Errorf("build in ADO failed, failed preparing ADO template parameters, err: %s", err)
	}

	fmt.Println("Running ADO pipeline.")
	adoClient := adopipelines.NewClient(o.ADOOrganizationURL, adoPAT)

	fmt.Println("Triggering ADO build pipeline")
	ctx := context.Background()
	pipelineRun, err := adopipelines.Run(ctx, adoClient, templateParameters, o.GetADOConfig())
	if err != nil {
		return fmt.Errorf("build in ADO failed, failed running ADO pipeline, err: %s", err)
	}

	pipelineRunResult, err := adopipelines.GetRunResult(ctx, adoClient, o.GetADOConfig(), pipelineRun.Id, 30*time.Second)
	if err != nil {
		return fmt.Errorf("build in ADO failed, failed getting ADO pipeline run result, err: %s", err)
	}
	fmt.Printf("ADO pipeline run finished with status: %s", *pipelineRunResult)

	fmt.Println("Getting ADO pipeline run logs.")
	adoBuildClient, err := adopipelines.NewBuildClient(o.ADOOrganizationURL, adoPAT)
	if err != nil {
		fmt.Printf("Can't read ADO pipeline run logs, failed creating ADO build client, err: %s", err)
	}
	logs, err := adopipelines.GetRunLogs(ctx, adoBuildClient, &http.Client{}, o.GetADOConfig(), pipelineRun.Id, adoPAT)
	if err != nil {
		fmt.Printf("Failed read ADO pipeline run logs, err: %s", err)
	} else {
		fmt.Printf("ADO pipeline image build logs:\n%s", logs)
	}

	if *pipelineRunResult == pipelines.RunResultValues.Failed || *pipelineRunResult == pipelines.RunResultValues.Unknown {
		return fmt.Errorf("build in ADO finished with status: %s", *pipelineRunResult)
	}
	return nil
}

func buildLocally(o options) error {
	runFunc := runInKaniko
	if os.Getenv("USE_BUILDKIT") == "true" {
		runFunc = runInBuildKit
	}
	var sha, pr string
	var err error
	var variant Variants
	var envs map[string]string

	context, err := filepath.Abs(o.Context)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// TODO(dekiel): validating if envFile or variants.yaml file exists should be done in validateOptions or in a separate function.
	// 		We should call this function before calling image building functions.
	dockerfilePath := filepath.Join(context, filepath.Dir(o.Dockerfile))

	if len(o.EnvFile) > 0 {
		envs, err = loadEnv(os.DirFS(dockerfilePath), o.EnvFile)
		if err != nil {
			return fmt.Errorf("load env failed, error: %w", err)
		}

	} else {
		variantsFile := filepath.Join(dockerfilePath, "variants.yaml")
		variant, err = GetVariants(o.Variant, variantsFile, os.ReadFile)
		if err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("get variants failed, error: %w", err)
			}
		}
		if len(variant) > 0 {
			return fmt.Errorf("building variants is not not working and is not supported anymore")
		}
	}

	repo := o.Config.Registry
	if o.IsCI {
		presubmit := os.Getenv("JOB_TYPE") == "presubmit"
		if presubmit {
			if len(o.DevRegistry) > 0 {
				repo = o.DevRegistry
			}
			if n := os.Getenv("PULL_NUMBER"); n != "" {
				pr = n
			}
		}

		if c := os.Getenv("PULL_BASE_SHA"); c != "" {
			sha = c
		}
	}

	// if sha is still not set, fail the pipeline
	if sha == "" {
		return fmt.Errorf("'sha' could not be determined")
	}

	parsedTags, err := getTags(pr, sha, append(o.Tags, o.TagTemplate))
	if err != nil {
		return err
	}

	// Provide parsedTags as buildArgs for developers
	var buildArgs map[string]string
	if o.ExportTags {
		buildArgs = addTagsToEnv(parsedTags, envs)
	} else {
		buildArgs = envs
	}

	if buildArgs == nil {
		buildArgs = make(map[string]string)
	}

	appendMissing(&buildArgs, o.BuildArgs)

	// variants.yaml file not present or either empty. Run single build.
	destinations := gatherDestinations(repo, o.Name, parsedTags)
	fmt.Println("Starting build for image: ", strings.Join(destinations, ", "))
	err = runFunc(o, "build", destinations, o.Platforms, buildArgs)
	if err != nil {
		return fmt.Errorf("build encountered error: %w", err)
	}

	err = signImages(&o, destinations)
	if err != nil {
		return fmt.Errorf("sign encountered error: %w", err)
	}
	fmt.Println("Successfully built image:", strings.Join(destinations, ", "))
	return nil
}

// appendMissing appends key, values pairs from source array to target map
func appendMissing(target *map[string]string, source []tags.Tag) {
	if len(source) > 0 {
		for _, arg := range source {
			if _, exists := (*target)[arg.Name]; !exists {
				(*target)[arg.Name] = arg.Value
			}
		}
	}
}

// TODO: write tests for this function
func signImages(o *options, images []string) error {
	// use o.OrgRepo as default value since someone might have loaded is as a flag
	orgRepo := o.OrgRepo
	if o.IsCI {
		// try to extract OrgRepo from Prow-based env variables
		org := os.Getenv("REPO_OWNER")
		repo := os.Getenv("REPO_NAME")
		if len(org) > 0 && len(repo) > 0 {
			// assume this is our variable since both variables are present
			orgRepo = org + "/" + repo
		}
	}
	if len(orgRepo) == 0 {
		return fmt.Errorf("'OrgRepo' cannot be empty")
	}
	sig, err := getSignersForOrgRepo(o, orgRepo)
	if err != nil {
		return err
	}
	fmt.Println("Start signing images", strings.Join(images, ","))
	var errs []error
	for _, s := range sig {
		err := s.Sign(images)
		if err != nil {
			errs = append(errs, fmt.Errorf("sign error: %w", err))
		}
	}
	return errutil.NewAggregate(errs)
}

// getSignersForOrgRepo fetches all signers for a repository
// It fetches all signers from '*' and specific org/repo combo.
func getSignersForOrgRepo(o *options, orgRepo string) ([]sign.Signer, error) {
	c := o.SignConfig
	if len(c.EnabledSigners) == 0 {
		// no signers enabled. no need to gather signers
		return nil, nil
	}
	var enabled StrList
	jobType := os.Getenv("JOB_TYPE")
	defaultSigners := c.EnabledSigners["*"]
	orgRepoSigners := c.EnabledSigners[orgRepo]
	for _, s := range append(defaultSigners, orgRepoSigners...) {
		enabled.Add(s)
	}
	fmt.Println("sign images using services", strings.Join(enabled.List(), ", "))
	var signers []sign.Signer
	for _, sc := range c.Signers {
		if enabled.Has(sc.Name) {
			// if signerConfig doesn't contain any jobTypes, it should be considered enabled by default
			if len(sc.JobType) > 0 && !o.IsCI {
				fmt.Println("signer", sc.Name, "ignored, because image-builder is not running in CI mode and contains 'job-type' field defined")
				continue
			}
			if len(jobType) > 0 && len(sc.JobType) > 0 && o.IsCI {
				var has bool
				for _, t := range sc.JobType {
					if t == jobType {
						has = true
						break
					}
				}
				if !has {
					// ignore signer if the jobType doesn't contain specific job type
					fmt.Println("signer", sc.Name, "ignored, because is not enabled for a CI job of type:", jobType)
					continue
				}
			}
			s, err := sc.Config.NewSigner()
			if err != nil {
				return nil, fmt.Errorf("signer init: %w", err)
			}
			signers = append(signers, s)
		}
	}
	return signers, nil
}

// StrList implements list of strings as a map
// This implementation allows getting unique values when merging multiple maps into one
// (@Ressetkk): We should find better place to move that code
type StrList struct {
	m map[string]interface{}
	sync.Mutex
}

func (l *StrList) Add(value string) {
	l.Lock()
	// lazy init map
	if l.m == nil {
		l.m = make(map[string]interface{})
	}
	if _, ok := l.m[value]; !ok {
		l.m[value] = new(interface{})
	}
	l.Unlock()
}

func (l *StrList) Has(elem string) bool {
	_, ok := l.m[elem]
	return ok
}

func (l *StrList) List() []string {
	var n []string
	for val := range l.m {
		n = append(n, val)
	}
	return n
}

func getTags(pr, sha string, templates []tags.Tag) ([]tags.Tag, error) {
	// (Ressetkk): PR tag should not be hardcoded, in the future we have to find a way to parametrize it
	if pr != "" {
		// assume we are using PR number, build default tag as 'PR-XXXX'
		return []tags.Tag{{Name: "default_tag", Value: "PR-" + pr}}, nil
	}
	// build a tag from commit SHA
	tagger, err := tags.NewTagger(templates, tags.CommitSHA(sha))
	if err != nil {
		return nil, fmt.Errorf("get tagger: %w", err)
	}
	p, err := tagger.ParseTags()
	if err != nil {
		return nil, fmt.Errorf("build tag: %w", err)
	}
	return p, nil
}

func gatherDestinations(repo []string, name string, tags []tags.Tag) []string {
	var dst []string
	for _, t := range tags {
		for _, r := range repo {
			image := path.Join(r, name)
			dst = append(dst, image+":"+strings.ReplaceAll(t.Value, " ", "-"))
		}
	}
	return dst
}

// validateOptions handles options validation. All checks should be provided here
func validateOptions(o options) error {
	var errs []error

	if o.Context == "" {
		errs = append(errs, fmt.Errorf("flag '--context' is missing"))
	}

	if o.Name == "" {
		errs = append(errs, fmt.Errorf("flag '--name' is missing"))
	}

	if o.Dockerfile == "" {
		errs = append(errs, fmt.Errorf("flag '--dockerfile' is missing"))
	}

	if o.ConfigPath == "" {
		errs = append(errs, fmt.Errorf("'--config' flag is missing or has empty value, please provide the path to valid 'config.yaml' file"))
	}

	if o.SignOnly && len(o.ImagesToSign) == 0 {
		errs = append(errs, fmt.Errorf("flag '--images-to-sign' is missing, please provide at least one image to sign"))
	}
	if !o.SignOnly && len(o.ImagesToSign) > 0 {
		errs = append(errs, fmt.Errorf("flag '--sign-only' is missing or has false value, please set it to true when using '--images-to-sign' flag"))
	}

	if o.EnvFile != "" && o.BuildInADO {
		errs = append(errs, fmt.Errorf("envFile flag is not supported when running in ADO"))
	}

	if o.Variant != "" && o.BuildInADO {
		errs = append(errs, fmt.Errorf("variant flag is not supported when running in ADO"))
	}

	return errutil.NewAggregate(errs)
}

// loadEnv loads environment variables into application runtime from key=value list
func loadEnv(vfs fs.FS, envFile string) (map[string]string, error) {
	if len(envFile) == 0 {
		// file is empty - ignore
		return nil, nil
	}
	f, err := vfs.Open(envFile)
	if err != nil {
		return nil, fmt.Errorf("open env file: %w", err)
	}
	s := bufio.NewScanner(f)
	vars := make(map[string]string)
	for s.Scan() {
		kv := s.Text()
		sp := strings.SplitN(kv, "=", 2)
		key, val := sp[0], sp[1]
		if len(sp) > 2 {
			return nil, fmt.Errorf("env var split incorrectly: 2 != %v", len(sp))
		}
		if _, ok := os.LookupEnv(key); ok {
			// do not override env variable if it's already present in the runtime
			// do not include in vars map since dev should not have access to it anyway
			continue
		}
		err := os.Setenv(key, val)
		if err != nil {
			return nil, fmt.Errorf("setenv: %w", err)
		}
		// add value to the vars that will be injected as build args
		vars[key] = val
	}
	return vars, nil
}

// Add parsed Tags to environments which will be passed to Dockerfile
func addTagsToEnv(tags []tags.Tag, envs map[string]string) map[string]string {
	m := make(map[string]string)

	for _, t := range tags {
		key := fmt.Sprintf("TAG_%s", t.Name)
		m[key] = t.Value
	}

	for k, v := range envs {
		m[k] = v
	}

	return m
}

func (o *options) gatherOptions(flagSet *flag.FlagSet) *flag.FlagSet {
	flagSet.BoolVar(&o.Silent, "silent", false, "Do not push build logs to stdout")
	flagSet.StringVar(&o.ConfigPath, "config", "/config/image-builder-config.yaml", "Path to application config file")
	flagSet.StringVar(&o.Context, "context", ".", "Path to build directory Context")
	flagSet.StringVar(&o.EnvFile, "env-file", "", "Path to file with environment variables to be loaded in build")
	flagSet.StringVar(&o.Name, "name", "", "Name of the image to be built")
	flagSet.StringVar(&o.Dockerfile, "dockerfile", "Dockerfile", "Path to Dockerfile file relative to Context")
	flagSet.StringVar(&o.Variant, "variant", "", "If variants.yaml file is present, define which Variant should be built. If variants.yaml is not present, this flag will be ignored")
	flagSet.StringVar(&o.LogDir, "log-dir", "/logs/artifacts", "Path to logs directory where GCB logs will be stored")
	// TODO: What is expected value repo only or org/repo? How this flag influence an image builder behaviour?
	flagSet.StringVar(&o.OrgRepo, "repo", "", "Load repository-specific configuration, for example, signing configuration")
	flagSet.Var(&o.Tags, "tag", "Additional tag that the image will be tagged with. Optionally you can pass the Name in the format Name=value which will be used by export-Tags")
	flagSet.Var(&o.BuildArgs, "build-arg", "Flag to pass additional arguments to build Dockerfile. It can be used in the Name=value format.")
	flagSet.Var(&o.Platforms, "platform", "Only supported with BuildKit. Platform of the image that is built")
	flagSet.BoolVar(&o.ExportTags, "export-Tags", false, "Export parsed Tags as build-args into Dockerfile. Each tag will have format TAG_x, where x is the tag Name passed along with the tag")
	flagSet.BoolVar(&o.SignOnly, "sign-only", false, "Only sign the image, do not build it")
	flagSet.Var(&o.ImagesToSign, "images-to-sign", "Comma-separated list of images to sign. Only used when sign-only flag is set")
	flagSet.BoolVar(&o.BuildInADO, "build-in-ado", false, "Build in Azure DevOps pipeline environment")
	flagSet.BoolVar(&o.ParseTagsOnly, "parse-tags-only", false, "Only parse tags and print them to stdout")

	return flagSet
}

func main() {
	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	o := options{IsCI: os.Getenv("CI") == "true"}
	o.gatherOptions(flagSet)
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// validate if options provided by flags and config file are fine
	if err := validateOptions(o); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	c, err := os.ReadFile(o.ConfigPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := o.ParseConfig(c); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if o.SignOnly {
		err = signImages(&o, o.ImagesToSign)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// TODO(dekiel): refactor this function to move all logic to separate function and make it testable.
	if o.ParseTagsOnly {
		var sha, pr string
		if o.IsCI {
			presubmit := os.Getenv("JOB_TYPE") == "presubmit"
			if presubmit {
				if n := os.Getenv("PULL_NUMBER"); n != "" {
					pr = n
				}
			}

			if c := os.Getenv("PULL_BASE_SHA"); c != "" {
				sha = c
			}
		}

		// if sha is still not set, fail the pipeline
		if sha == "" {
			fmt.Println("'sha' could not be determined")
			os.Exit(1)
		}
		parsedTags, err := getTags(pr, sha, append(o.Tags, o.TagTemplate))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// Print parsed tags to stdout as json
		jsonTags, err := json.Marshal(parsedTags)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("%s", jsonTags)
		os.Exit(0)
	}
	if o.BuildInADO {
		err = buildInADO(o)
		if err != nil {
			fmt.Printf("Image build failed with error: %s", err)
			os.Exit(1)
		}
	}
	err = buildLocally(o)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Job's done.")
}
