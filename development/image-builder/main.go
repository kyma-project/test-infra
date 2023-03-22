package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/kyma-project/test-infra/development/image-builder/sign"
	"github.com/kyma-project/test-infra/development/pkg/sets"
	"github.com/kyma-project/test-infra/development/pkg/tags"
	errutil "k8s.io/apimachinery/pkg/util/errors"
)

type options struct {
	Config
	configPath string
	context    string
	dockerfile string
	envFile    string
	name       string
	variant    string
	logDir     string
	orgRepo    string
	silent     bool
	isCI       bool
	tags       sets.Tags
	buildArgs  sets.Tags
	platforms  sets.Strings
	exportTags bool
}

const (
	PlatformLinuxAmd64 = "linux/amd64"
	PlatformLinuxArm64 = "linux/arm64"
)

// parseVariable returns a build-arg.
// Keys are set to upper-case.
func parseVariable(key, value string) string {
	k := strings.TrimSpace(key)
	return k + "=" + strings.TrimSpace(value)
}

// runInBuildKit prepares command execution and handles gathering logs from BuildKit-enabled run
// This function is used only in customized environment
func runInBuildKit(o options, name string, destinations, platforms []string, buildArgs map[string]string) error {
	dockerfile := filepath.Base(o.dockerfile)
	dockerfileDir := filepath.Dir(o.dockerfile)
	args := []string{
		"build", "--frontend=dockerfile.v0",
		"--local", "context=" + o.context,
		"--local", "dockerfile=" + filepath.Join(o.context, dockerfileDir),
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

	if !o.silent {
		outw = append(outw, os.Stdout)
		errw = append(errw, os.Stderr)
	}

	f, err := os.Create(filepath.Join(o.logDir, strings.TrimSpace("build_"+strings.TrimSpace(name)+".log")))
	if err != nil {
		return fmt.Errorf("could not create log file: %w", err)
	}

	outw = append(outw, f)
	errw = append(errw, f)

	cmd.Stdout = io.MultiWriter(outw...)
	cmd.Stderr = io.MultiWriter(errw...)

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// runInKaniko prepares command execution and handles gathering logs to file
func runInKaniko(o options, name string, destinations, platforms []string, buildArgs map[string]string) error {
	args := []string{
		"--context=" + o.context,
		"--dockerfile=" + o.dockerfile,
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

	if !o.silent {
		outw = append(outw, os.Stdout)
		errw = append(errw, os.Stderr)
	}

	f, err := os.Create(filepath.Join(o.logDir, strings.TrimSpace("build_"+strings.TrimSpace(name)+".log")))
	if err != nil {
		return fmt.Errorf("could not create log file: %w", err)
	}

	outw = append(outw, f)
	errw = append(errw, f)

	cmd.Stdout = io.MultiWriter(outw...)
	cmd.Stderr = io.MultiWriter(errw...)

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func runBuildJob(o options, vs Variants, envs map[string]string) error {
	runFunc := runInKaniko
	if os.Getenv("USE_BUILDKIT") == "true" {
		runFunc = runInBuildKit
	}
	var sha, pr string
	var err error
	repo := o.Config.Registry
	if o.isCI {
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

	parsedTags, err := getTags(pr, sha, append(o.tags, o.TagTemplate))
	if err != nil {
		return err
	}

	// Provide parsedTags as buildArgs for developers
	var buildArgs map[string]string
	if o.exportTags {
		buildArgs = addTagsToEnv(parsedTags, envs)
	} else {
		buildArgs = envs
	}

	if buildArgs == nil {
		buildArgs = make(map[string]string)
	}

	appendMissing(&buildArgs, o.buildArgs)

	if len(vs) == 0 {
		// variants.yaml file not present or either empty. Run single build.
		destinations := gatherDestinations(repo, o.name, parsedTags)
		fmt.Println("Starting build for image: ", strings.Join(destinations, ", "))
		err = runFunc(o, "build", destinations, o.platforms, buildArgs)
		if err != nil {
			return fmt.Errorf("build encountered error: %w", err)
		}

		err := signImages(&o, destinations)
		if err != nil {
			return fmt.Errorf("sign encountered error: %w", err)
		}
		fmt.Println("Successfully built image:", strings.Join(destinations, ", "))
		return nil
	}
	return fmt.Errorf("building variants is not supported at this moment")
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

func signImages(o *options, images []string) error {
	// use o.orgRepo as default value since someone might have loaded is as a flag
	orgRepo := o.orgRepo
	if o.isCI {
		// try to extract orgRepo from Prow-based env variables
		org := os.Getenv("REPO_OWNER")
		repo := os.Getenv("REPO_NAME")
		if len(org) > 0 && len(repo) > 0 {
			// assume this is our variable since both variables are present
			orgRepo = org + "/" + repo
		}
	}
	if len(orgRepo) == 0 {
		return fmt.Errorf("'orgRepo' cannot be empty")
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
			if len(sc.JobType) > 0 && !o.isCI {
				fmt.Println("signer", sc.Name, "ignored, because image-builder is not running in CI mode and contains 'job-type' field defined")
				continue
			}
			if len(jobType) > 0 && len(sc.JobType) > 0 && o.isCI {
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
	if o.context == "" {
		errs = append(errs, fmt.Errorf("flag '--context' is missing"))
	}
	if o.name == "" {
		errs = append(errs, fmt.Errorf("flag '--name' is missing"))
	}
	if o.dockerfile == "" {
		errs = append(errs, fmt.Errorf("flag '--dockerfile' is missing"))
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

// Add parsed tags to environments which will be passed to dockerfile
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
	flagSet.BoolVar(&o.silent, "silent", false, "Do not push build logs to stdout")
	flagSet.StringVar(&o.configPath, "config", "/config/image-builder-config.yaml", "Path to application config file")
	flagSet.StringVar(&o.context, "context", ".", "Path to build directory context")
	flagSet.StringVar(&o.envFile, "env-file", "", "Path to file with environment variables to be loaded in build")
	flagSet.StringVar(&o.name, "name", "", "Name of the image to be built")
	flagSet.StringVar(&o.dockerfile, "dockerfile", "Dockerfile", "Path to Dockerfile file relative to context")
	flagSet.StringVar(&o.variant, "variant", "", "If variants.yaml file is present, define which variant should be built. If variants.yaml is not present, this flag will be ignored")
	flagSet.StringVar(&o.logDir, "log-dir", "/logs/artifacts", "Path to logs directory where GCB logs will be stored")
	flagSet.StringVar(&o.orgRepo, "repo", "", "Load repository-specific configuration, for example, signing configuration")
	flagSet.Var(&o.tags, "tag", "Additional tag that the image will be tagged with. Optionally you can pass the name in the format name=value which will be used by export-tags")
	flagSet.Var(&o.buildArgs, "build-arg", "Flag to pass additional arguments to build Dockerfile. It can be used in the name=value format.")
	flagSet.Var(&o.platforms, "platform", "Only supported with BuildKit. Platform of the image that is built")
	flagSet.BoolVar(&o.exportTags, "export-tags", false, "Export parsed tags as build-args into Dockerfile. Each tag will have format TAG_x, where x is the tag name passed along with the tag")
	return flagSet
}

func main() {
	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	o := options{isCI: os.Getenv("CI") == "true"}
	o.gatherOptions(flagSet)
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if o.configPath == "" {
		fmt.Println("'--config' flag is missing or has empty value, please provide the path to valid 'config.yaml' file")
		os.Exit(1)
	}
	c, err := os.ReadFile(o.configPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := o.ParseConfig(c); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// validate if options provided by flags and config file are fine
	if err := validateOptions(o); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	context, err := filepath.Abs(o.context)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	dockerfilePath := filepath.Join(context, filepath.Dir(o.dockerfile))

	var variant Variants
	var envs map[string]string
	if len(o.envFile) > 0 {
		envs, err = loadEnv(os.DirFS(dockerfilePath), o.envFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	} else {
		variantsFile := filepath.Join(dockerfilePath, "variants.yaml")
		variant, err = GetVariants(o.variant, variantsFile, os.ReadFile)
		if err != nil {
			if !os.IsNotExist(err) {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}

	err = runBuildJob(o, variant, envs)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Job's done.")
}
