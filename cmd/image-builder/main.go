package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	adopipelines "github.com/kyma-project/test-infra/pkg/azuredevops/pipelines"
	"github.com/kyma-project/test-infra/pkg/extractimageurls"
	"github.com/kyma-project/test-infra/pkg/github/actions"
	"github.com/kyma-project/test-infra/pkg/sets"
	"github.com/kyma-project/test-infra/pkg/sign"
	"github.com/kyma-project/test-infra/pkg/tags"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/pipelines"
	"golang.org/x/net/context"
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
	tagsBase64 string
	buildArgs  sets.Tags
	platforms  sets.Strings
	exportTags bool
	// signOnly only sign images. No build will be performed.
	signOnly              bool
	imagesToSign          sets.Strings
	buildInADO            bool
	adoPreviewRun         bool
	adoPreviewRunYamlPath string
	parseTagsOnly         bool
	oidcToken             string
	azureAccessToken      string
	ciSystem              CISystem
	gitState              GitStateConfig
	debug                 bool
	dryRun                bool
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

	return cmd.Run()
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

	return cmd.Run()
}

// prepareADOTemplateParameters is a function that prepares the parameters for the Azure DevOps oci-image-builder pipeline.
// These parameters are used to trigger the pipeline with API call and build the image in the ADO environment.
// It takes an options struct as an argument and returns an OCIImageBuilderTemplateParams struct and an error.
// The function fetches various environment variables such as REPO_NAME, REPO_OWNER, JOB_TYPE, PULL_NUMBER, PULL_BASE_SHA, and PULL_PULL_SHA.
// It validates these variables are present and sets them in the templateParameters struct.
// It also sets other parameters from the options struct such as imageName, dockerfilePath, buildContext, exportTags, useKanikoConfigFromPR, buildArgs, and imageTags.
// The function validates the templateParameters and returns it along with any error that occurred during the process.
// TODO: rename this function to indicate that it's preparing ADO pipeline parameters for oci-image-builder pipeline.
func prepareADOTemplateParameters(options options) (adopipelines.OCIImageBuilderTemplateParams, error) {
	templateParameters := make(adopipelines.OCIImageBuilderTemplateParams)

	templateParameters.SetRepoName(options.gitState.RepositoryName)

	templateParameters.SetRepoOwner(options.gitState.RepositoryOwner)

	switch options.gitState.JobType {
	case "presubmit":
		templateParameters.SetPresubmitJobType()
	case "postsubmit":
		templateParameters.SetPostsubmitJobType()
	case "workflow_dispatch":
		templateParameters.SetWorkflowDispatchJobType()
	case "schedule":
		templateParameters.SetScheduleJobType()
	default:
		return nil, fmt.Errorf("unknown JobType received, ensure image-builder runs on supported event")
	}

	if options.gitState.IsPullRequest() {
		templateParameters.SetPullNumber(fmt.Sprint(options.gitState.PullRequestNumber))
	}

	templateParameters.SetBaseSHA(options.gitState.BaseCommitSHA)

	if len(options.gitState.BaseCommitRef) > 0 {
		templateParameters.SetBaseRef(options.gitState.BaseCommitRef)
	}

	if options.gitState.IsPullRequest() {
		templateParameters.SetPullSHA(options.gitState.PullHeadCommitSHA)
	}

	templateParameters.SetImageName(options.name)

	templateParameters.SetDockerfilePath(options.dockerfile)

	if len(options.envFile) > 0 {
		templateParameters.SetEnvFilePath(options.envFile)
	}

	templateParameters.SetBuildContext(options.context)

	templateParameters.SetExportTags(options.exportTags)

	if len(options.buildArgs) > 0 {
		templateParameters.SetBuildArgs(options.buildArgs.String())
	}

	if len(options.tags) > 0 {
		templateParameters.SetImageTags(options.tags.String())
	}

	if options.ciSystem == GithubActions {
		templateParameters.SetAuthorization(options.oidcToken)
	}

	err := templateParameters.Validate()
	if err != nil {
		return nil, fmt.Errorf("failed validating ADO template parameters, err: %w", err)
	}

	return templateParameters, nil
}

// buildInADO is a function that triggers the Azure DevOps (ADO) pipeline to build an image.
// It takes an options struct as an argument and returns an error.
// The function fetches the ADO_PAT environment variable and validates it's present.
// ADO_PAT holds personal access token and is used to authenticate with the ADO API.
// The function prepares the ADO pipeline parameters by calling the prepareADOTemplateParameters function.
// It creates a new ADO client and prepares the ADO pipeline run arguments.
// The function triggers the ADO build pipeline and waits for the pipeline run to finish.
// It fetches the ADO pipeline run logs and prints them.
// The function can trigger the ADO pipeline in preview mode if the adoPreviewRun flag is set to true.
// In preview mode, the function prints the final yaml of the ADO pipeline run.
// Running in preview mode requires the adoPreviewRunYamlPath flag to be set to the path of the yaml file with the ADO pipeline definition.
// This is used for pipeline syntax validation.
// If the pipeline run fails, the function returns an error.
// If the pipeline run is successful, the function returns nil.
// TODO(dekiel): refactor this function to accept clients as parameters to make it testable with mocks.
func buildInADO(o options) error {
	fmt.Println("Building image in ADO pipeline.")

	// Getting Azure DevOps Personal Access Token (ADO_PAT) from environment variable for authentication with ADO API when it's not set via flag.
	if o.azureAccessToken == "" && !o.dryRun {
		adoPAT, present := os.LookupEnv("ADO_PAT")
		if !present {
			return fmt.Errorf("build in ADO failed, ADO_PAT environment variable is not set, please set it to valid ADO PAT")
		}
		o.azureAccessToken = adoPAT
	} else if o.dryRun {
		fmt.Println("Running in dry-run mode. Skipping getting ADO PAT.")
	}

	fmt.Println("Preparing ADO template parameters.")
	// Preparing ADO pipeline parameters.
	templateParameters, err := prepareADOTemplateParameters(o)
	if err != nil {
		return fmt.Errorf("build in ADO failed, failed preparing ADO template parameters, err: %s", err)
	}
	fmt.Printf("Using TemplateParameters: %+v\n", templateParameters)

	// Creating a new ADO pipelines client.
	adoClient := adopipelines.NewClient(o.AdoConfig.ADOOrganizationURL, o.azureAccessToken)

	var opts []adopipelines.RunPipelineArgsOptions
	// If running in preview mode, add a preview run option to the ADO pipeline run arguments.
	if o.adoPreviewRun {
		fmt.Println("Running in preview mode.")
		// Adding a path to the yaml file with the ADO pipeline definition for parsing it in a preview run.
		opts = append(opts, adopipelines.PipelinePreviewRun(o.adoPreviewRunYamlPath))
	}

	fmt.Println("Preparing ADO pipeline run arguments.")
	// Composing ADO pipeline run arguments.
	runPipelineArgs, err := adopipelines.NewRunPipelineArgs(templateParameters, o.AdoConfig.GetADOConfig(), opts...)
	if err != nil {
		return fmt.Errorf("build in ADO failed, failed creating ADO pipeline run args, err: %s", err)
	}

	fmt.Println("Triggering ADO build pipeline")
	var (
		pipelineRunResult *pipelines.RunResult
		logs              string
	)
	if !o.dryRun {
		ctx := context.Background()
		// Triggering ADO build pipeline.
		pipelineRun, err := adoClient.RunPipeline(ctx, runPipelineArgs)
		if err != nil {
			return fmt.Errorf("build in ADO failed, failed running ADO pipeline, err: %s", err)
		}

		// If running in preview mode, print the final yaml of ADO pipeline run for provided ADO pipeline definition and return.
		if o.adoPreviewRun {
			if pipelineRun.FinalYaml != nil {
				fmt.Printf("ADO pipeline preview run final yaml\n: %s", *pipelineRun.FinalYaml)
			} else {
				fmt.Println("ADO pipeline preview run final yaml is empty")
			}
			return nil
		}

		// Fetch the ADO pipeline run result.
		// GetRunResult function waits for the pipeline runs to finish and returns the result.
		// TODO(dekiel) make the timeout configurable instead of hardcoding it.
		pipelineRunResult, err = adopipelines.GetRunResult(ctx, adoClient, o.AdoConfig.GetADOConfig(), pipelineRun.Id)
		if err != nil {
			return fmt.Errorf("build in ADO failed, failed getting ADO pipeline run result, err: %s", err)
		}
		fmt.Printf("ADO pipeline run finished with status: %s\n", *pipelineRunResult)

		// Fetch the ADO pipeline run logs.
		fmt.Println("Getting ADO pipeline run logs.")
		// Creating a new ADO build client.
		adoBuildClient, err := adopipelines.NewBuildClient(o.AdoConfig.ADOOrganizationURL, o.azureAccessToken)
		if err != nil {
			fmt.Printf("Can't read ADO pipeline run logs, failed creating ADO build client, err: %s", err)
		}
		logs, err = adopipelines.GetRunLogs(ctx, adoBuildClient, &http.Client{}, o.AdoConfig.GetADOConfig(), pipelineRun.Id, o.azureAccessToken)
		if err != nil {
			fmt.Printf("Failed read ADO pipeline run logs, err: %s", err)
		} else {
			fmt.Printf("ADO pipeline image build logs:\n%s", logs)
		}
	} else {
		dryRunPipelineRunResult := pipelines.RunResult("Succeeded")
		pipelineRunResult = &dryRunPipelineRunResult
	}

	// TODO: Setting github outputs should happen outside buildInADO function.
	//  buildInADO should return required data and caller should handle it.
	// if run in github actions, set output parameters
	if o.ciSystem == GithubActions {
		fmt.Println("Setting GitHub outputs.")
		var images []string
		if !o.dryRun {
			images = extractImagesFromADOLogs(logs)
			fmt.Printf("Extracted built images from ADO logs: %v\n", images)
		} else {
			fmt.Println("Running in dry-run mode. Skipping extracting images and results from ADO.")
			images = []string{"registry/repo/image1:tag1", "registry/repo/image2:tag2"}
		}
		data, err := json.Marshal(images)
		if err != nil {
			return fmt.Errorf("cannot marshal list of images: %w", err)
		}

		err = actions.SetOutput("images", string(data))
		if err != nil {
			return fmt.Errorf("cannot set images GitHub output: %w", err)
		}
		fmt.Println("images GitHub output set")
		err = actions.SetOutput("adoResult", string(*pipelineRunResult))
		if err != nil {
			return fmt.Errorf("cannot set adoResult GitHub output: %w", err)
		}
		fmt.Println("adoResult GitHub output set")
	}

	// Handle the ADO pipeline run failure.
	if *pipelineRunResult == pipelines.RunResultValues.Failed || *pipelineRunResult == pipelines.RunResultValues.Unknown {
		return fmt.Errorf("build in ADO finished with status: %s", *pipelineRunResult)
	}
	return nil
}

// buildLocally is a function that builds an image locally using either Kaniko or BuildKit.
// It takes an options struct as an argument and returns an error.
// The function determines the build tool to use based on the USE_BUILDKIT environment variable.
// If USE_BUILDKIT is set to "true", BuildKit is used, otherwise Kaniko is used.
// The function fetches various environment variables such as JOB_TYPE, PULL_NUMBER, and PULL_BASE_SHA.
// It validates these variables are present and sets them in the appropriate variables.
// It also sets other variables from the options struct such as context, envFile, name, dockerfile, variant, and tags.
// The function validates the options and returns an error if any of the required options are not set or are invalid.
// The function triggers the build process and waits for it to finish.
// It fetches the build logs and prints them.
// If the build fails, the function returns an error.
// If the build is successful, the function returns nil.
func buildLocally(o options) error {
	// Determine the build tool to use based on the USE_BUILDKIT environment variable.
	runFunc := runInKaniko
	if os.Getenv("USE_BUILDKIT") == "true" {
		runFunc = runInBuildKit
	}
	var sha, pr string
	var err error
	var variant Variants
	var envs map[string]string

	// TODO(dekiel): validating if envFile or variants.yaml file exists should be done in validateOptions or in a separate function.
	// 		We should call this function before calling image building functions.
	dockerfilePath, err := getDockerfileDirPath(o)
	if err != nil {
		return fmt.Errorf("get dockerfile path failed, error: %w", err)
	}
	// Load environment variables from the envFile or variants.yaml file.
	if len(o.envFile) > 0 {
		envs, err = loadEnv(os.DirFS(dockerfilePath), o.envFile)
		if err != nil {
			return fmt.Errorf("load env failed, error: %w", err)
		}

	} else {
		variantsFile := filepath.Join(dockerfilePath, "variants.yaml")
		variant, err = GetVariants(o.variant, variantsFile, os.ReadFile)
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

	defaultTag, err := getDefaultTag(o)
	if err != nil {
		return err
	}

	// Get the tags for the image.
	parsedTags, err := getTags(pr, sha, append(o.tags, defaultTag))
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

	// variants.yaml file not present or either empty. Run single build.
	destinations := gatherDestinations(repo, o.name, parsedTags)
	fmt.Println("Starting build for image: ", strings.Join(destinations, ", "))
	err = runFunc(o, "build", destinations, o.platforms, buildArgs)
	if err != nil {
		return fmt.Errorf("build encountered error: %w", err)
	}

	// Sign the images.
	err = signImages(&o, destinations)
	if err != nil {
		return fmt.Errorf("sign encountered error: %w", err)
	}
	fmt.Println("Successfully built image:", strings.Join(destinations, ", "))
	return nil
}

// appendMissing appends key, values pairs from source array to target map
func appendMissing(target *map[string]string, source []tags.Tag) {
	for _, arg := range source {
		if _, exists := (*target)[arg.Name]; !exists {
			(*target)[arg.Name] = arg.Value
		}
	}
}

// appendToTags appends key-value pairs from source map to target slice of tags.Tag
// This allows creation of image tags from key value pairs.
func appendToTags(target *[]tags.Tag, source map[string]string) {
	for key, value := range source {
		*target = append(*target, tags.Tag{Name: key, Value: value})
	}
}

// TODO: write tests for this function
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
	var taggerOptions []tags.TagOption
	if len(pr) > 0 {
		taggerOptions = append(taggerOptions, tags.PRNumber(pr))
	}
	if len(sha) > 0 {
		taggerOptions = append(taggerOptions, tags.CommitSHA(sha))
	}

	// build a tag from commit SHA
	tagger, err := tags.NewTagger(templates, taggerOptions...)
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

	if o.configPath == "" {
		errs = append(errs, fmt.Errorf("'--config' flag is missing or has empty value, please provide the path to valid 'config.yaml' file"))
	}

	if o.signOnly && len(o.imagesToSign) == 0 {
		errs = append(errs, fmt.Errorf("flag '--images-to-sign' is missing, please provide at least one image to sign"))
	}
	if !o.signOnly && len(o.imagesToSign) > 0 {
		errs = append(errs, fmt.Errorf("flag '--sign-only' is missing or has false value, please set it to true when using '--images-to-sign' flag"))
	}

	if o.variant != "" && o.buildInADO {
		errs = append(errs, fmt.Errorf("variant flag is not supported when running in ADO"))
	}

	if o.adoPreviewRun && !o.buildInADO {
		errs = append(errs, fmt.Errorf("ado-preview-run flag is not supported when running locally"))
	}

	if o.adoPreviewRun && o.adoPreviewRunYamlPath == "" {
		errs = append(errs, fmt.Errorf("ado-preview-run-yaml-path flag is missing, please provide path to yaml file with ADO pipeline definition"))
	}

	if o.adoPreviewRunYamlPath != "" && !o.adoPreviewRun {
		errs = append(errs, fmt.Errorf("ado-preview-run-yaml-path flag is provided, but adoPreviewRun flag is not set to true"))
	}

	return errutil.NewAggregate(errs)
}

// loadEnv creates environment variables in application runtime from a file with key=value data
func loadEnv(vfs fs.FS, envFile string) (map[string]string, error) {
	if len(envFile) == 0 {
		// file is empty - ignore
		return nil, nil
	}
	file, err := vfs.Open(envFile)
	if err != nil {
		return nil, fmt.Errorf("open env file: %w", err)
	}
	fileReader := bufio.NewScanner(file)
	vars := make(map[string]string)
	for fileReader.Scan() {
		line := fileReader.Text()
		separatedValues := strings.SplitN(line, "=", 2)
		if len(separatedValues) > 2 {
			return nil, fmt.Errorf("env var split incorrectly, got more than two values, expected only two, values: %v", separatedValues)
		}
		// ignore empty lines, setup environment variable only if key and value are present
		if len(separatedValues) == 2 {
			key, val := separatedValues[0], separatedValues[1]
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
	flagSet.StringVar(&o.name, "name", "", "name of the image to be built")
	flagSet.StringVar(&o.dockerfile, "dockerfile", "dockerfile", "Path to dockerfile file relative to context")
	flagSet.StringVar(&o.variant, "variant", "", "If variants.yaml file is present, define which variant should be built. If variants.yaml is not present, this flag will be ignored")
	flagSet.StringVar(&o.logDir, "log-dir", "/logs/artifacts", "Path to logs directory where GCB logs will be stored")
	flagSet.BoolVar(&o.debug, "debug", false, "Enable debug logging")
	flagSet.BoolVar(&o.dryRun, "dry-run", false, "Do not build the image, only print a ADO API call pipeline parameters")
	// TODO: What is expected value repo only or org/repo? How this flag influence an image builder behaviour?
	flagSet.StringVar(&o.orgRepo, "repo", "", "Load repository-specific configuration, for example, signing configuration")
	flagSet.Var(&o.tags, "tag", "Additional tag that the image will be tagged with. Optionally you can pass the name in the format name=value which will be used by export-tags")
	flagSet.StringVar(&o.tagsBase64, "tag-base64", "", "String representation of all tags encoded by base64. String representation must be in format as output of kyma-project/test-infra/pkg/tags.Tags.String() method")
	flagSet.Var(&o.buildArgs, "build-arg", "Flag to pass additional arguments to build dockerfile. It can be used in the name=value format.")
	flagSet.Var(&o.platforms, "platform", "Only supported with BuildKit. Platform of the image that is built")
	flagSet.BoolVar(&o.exportTags, "export-tags", false, "Export parsed tags as build-args into dockerfile. Each tag will have format TAG_x, where x is the tag name passed along with the tag")
	flagSet.BoolVar(&o.signOnly, "sign-only", false, "Only sign the image, do not build it")
	flagSet.Var(&o.imagesToSign, "images-to-sign", "Comma-separated list of images to sign. Only used when sign-only flag is set")
	flagSet.BoolVar(&o.buildInADO, "build-in-ado", false, "Build in Azure DevOps pipeline environment")
	flagSet.BoolVar(&o.adoPreviewRun, "ado-preview-run", false, "Trigger ADO pipeline in preview mode")
	flagSet.StringVar(&o.adoPreviewRunYamlPath, "ado-preview-run-yaml-path", "", "Path to yaml file with ADO pipeline definition to be used in preview mode")
	flagSet.BoolVar(&o.parseTagsOnly, "parse-tags-only", false, "Only parse tags and print them to stdout")
	flagSet.StringVar(&o.oidcToken, "oidc-token", "", "Token used to authenticate against Azure DevOps backend service")
	flagSet.StringVar(&o.azureAccessToken, "azure-access-token", "", "Token used to authenticate against Azure DevOps API")

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

	// If running inside some CI system, determine which system is used
	if o.isCI {
		ciSystem, err := DetermineUsedCISystem()
		if err != nil {
			log.Fatalf("Failed to determine current ci system: %s", err)
		}
		o.ciSystem = ciSystem
		o.gitState, err = LoadGitStateConfig(ciSystem)
		if err != nil {
			log.Fatalf("Failed to load current git state: %s", err)
		}
	}

	// validate if options provided by flags and config file are fine
	if err := validateOptions(o); err != nil {
		fmt.Println(err)
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

	if o.signOnly {
		err = signImages(&o, o.imagesToSign)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if o.parseTagsOnly {
		err = generateTags(o)
		if err != nil {
			fmt.Printf("Parse tags failed with error: %s\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	if o.buildInADO {
		err = buildInADO(o)
		if err != nil {
			fmt.Printf("Image build failed with error: %s\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	err = buildLocally(o)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Job's done.")
}

func generateTags(o options) error {
	// Get the absolute path to the dockerfile directory.
	dockerfilePath, err := getDockerfileDirPath(o)
	if err != nil {
		return fmt.Errorf("failed to get dockerfile path: %s", err)
	}
	// Load environment variables from the envFile.
	envs, err := getEnvs(o, dockerfilePath)
	if err != nil {
		return err
	}
	// Parse tags from the provided options.
	parsedTags, err := parseTags(o)
	if err != nil {
		return fmt.Errorf("failed to parse tags : %s", err)
	}
	// Append environment variables to tags.
	appendToTags(&parsedTags, envs)
	// Print parsed tags to stdout as json.
	jsonTags := tagsAsJSON(parsedTags)
	fmt.Printf("%s\n", jsonTags)
	return nil
}

func tagsAsJSON(parsedTags []tags.Tag) string {
	jsonTags, err := json.Marshal(parsedTags)
	if err != nil {
		fmt.Printf("Failed to marshal tags to json: %s", err)
		os.Exit(1)
	}
	return string(jsonTags)
}

func getEnvs(o options, dockerfilePath string) (map[string]string, error) {
	if len(o.envFile) > 0 {
		envs, err := loadEnv(os.DirFS(dockerfilePath), o.envFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load env file: %s", err)
		}
		return envs, nil
	}
	return map[string]string{}, nil
}

func parseTags(o options) ([]tags.Tag, error) {
	var (
		pr  string
		sha string
	)
	if !o.gitState.isPullRequest && o.gitState.BaseCommitSHA != "" {
		sha = o.gitState.BaseCommitSHA
	}
	if o.gitState.isPullRequest && o.gitState.PullRequestNumber > 0 {
		pr = fmt.Sprint(o.gitState.PullRequestNumber)
	}

	// TODO (dekiel): Tags provided as base64 encoded string should be parsed and added to the tags list when parsing flags.
	//   This way all tags are available in the tags list from thr very beginning of execution and can be used in any process.
	// read tags from base64 encoded string if provided
	if o.tagsBase64 != "" {
		decoded, err := base64.StdEncoding.DecodeString(o.tagsBase64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode tags, error: %w", err)
		}
		splitedTags := strings.Split(string(decoded), ",")
		for _, tag := range splitedTags {
			err = o.tags.Set(tag)
			if err != nil {
				return nil, fmt.Errorf("failed to set tag, tag: %s, error: %w", tag, err)
			}
		}
	}

	defaultTag, err := getDefaultTag(o)
	if err != nil {
		return nil, err
	}
	parsedTags, err := getTags(pr, sha, append(o.tags, defaultTag))
	if err != nil {
		return nil, err
	}

	return parsedTags, nil
}

// getDefaultTag returns the default tag based on the read git state.
// The function provid default tag for pull request or commit.
// The default tag is read from the provided options struct.
func getDefaultTag(o options) (tags.Tag, error) {
	if o.gitState.isPullRequest && o.gitState.PullRequestNumber > 0 {
		return o.DefaultPRTag, nil
	}
	if len(o.gitState.BaseCommitSHA) > 0 {
		return o.DefaultCommitTag, nil
	}
	return tags.Tag{}, fmt.Errorf("could not determine default tag, no pr number or commit sha provided")
}

func getDockerfileDirPath(o options) (string, error) {
	// Get the absolute path to the build context directory.
	context, err := filepath.Abs(o.context)
	if err != nil {
		return "", fmt.Errorf("could not get absolute path to build context directory: %w", err)
	}
	// Get the absolute path to the dockerfile.
	dockerfileDirPath := filepath.Join(context, filepath.Dir(o.dockerfile))
	return dockerfileDirPath, err
}

// extractImagesFromADOLogs extract docker images from Azure DevOps logs to allow us prepare list of images built in ADO backend
// The list can be than saved and provided as input for developers to use in next steps of their workflows.
// ADO Logs that we fetch anyway are the simplest solution to get such list from ADO backend.
func extractImagesFromADOLogs(logs string) []string {
	re := regexp.MustCompile(`--images-to-sign=(([a-z0-9]+(?:[.-][a-z0-9]+)*/)*([a-z0-9]+(?:[.-][a-z0-9]+)*)(?::[a-z0-9.-]+)?/([a-z0-9-]+)/([a-z0-9-]+)(?::[a-zA-Z0-9.-]+))`)
	matches := re.FindAllStringSubmatch(logs, -1)

	images := []string{}
	if len(matches) > 0 {
		for _, match := range matches {
			if len(match) > 1 {
				images = append(images, match[1])
			}
		}
	}

	images = extractimageurls.UniqueImages(images)

	return images
}
