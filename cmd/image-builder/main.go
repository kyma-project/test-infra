package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	adopipelines "github.com/kyma-project/test-infra/pkg/azuredevops/pipelines"
	"github.com/kyma-project/test-infra/pkg/github/actions"
	"github.com/kyma-project/test-infra/pkg/imagebuilder"
	"github.com/kyma-project/test-infra/pkg/logging"
	"github.com/kyma-project/test-infra/pkg/sets"
	"github.com/kyma-project/test-infra/pkg/sign"
	"github.com/kyma-project/test-infra/pkg/tags"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/pipelines"
	"go.uber.org/zap"
	errutil "k8s.io/apimachinery/pkg/util/errors"
)

type options struct {
	Config
	configPath string
	context    string
	dockerfile string
	name       string
	logDir     string
	logger     Logger
	orgRepo    string
	silent     bool
	isCI       bool
	tags       sets.Tags
	tagsBase64 string
	buildArgs  sets.Tags
	platforms  sets.Strings
	exportTags bool
	// signOnly only sign images. No build will be performed.
	signOnly                bool
	imagesToSign            sets.Strings
	adoPreviewRun           bool
	adoPreviewRunYamlPath   string
	parseTagsOnly           bool
	oidcToken               string
	azureAccessToken        string
	ciSystem                CISystem
	gitState                GitStateConfig
	debug                   bool
	dryRun                  bool
	tagsOutputFile          string
	useGoInternalSAPModules bool
	// buildReportPath is a path to the file where the build report will be saved
	// build report will be used by SRE team to gather information about the build
	buildReportPath string
	// adoStateOutput indicates if the success or failure of the command (sign or build) should be
	// reported as an output variable in Azure DevOps
	adoStateOutput bool
	target         string
}

type Logger interface {
	logging.StructuredLoggerInterface
	logging.WithLoggerInterface
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
	case "merge_group":
		templateParameters.SetMergeGroupJobType()
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

	templateParameters.SetBuildContext(options.context)

	templateParameters.SetExportTags(options.exportTags)

	if len(options.buildArgs) > 0 {
		templateParameters.SetBuildArgs(options.buildArgs.String())
	}

	if len(options.tags) > 0 {
		templateParameters.SetImageTags(options.tags.String())
	}

	if options.oidcToken != "" {
		templateParameters.SetAuthorization(options.oidcToken)
	}

	if options.useGoInternalSAPModules {
		templateParameters.SetUseGoInternalSAPModules()
	}

	if len(options.platforms) > 0 {
		templateParameters.SetPlatforms(options.platforms.String())
	} else {
		// Set default platforms to linux/amd64,linux/arm64, if not set. There is no way to set during flag parsing.
		templateParameters.SetPlatforms("linux/amd64,linux/arm64")
	}

	if options.target != "" {
		templateParameters.SetTarget(options.target)
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
		buildReport       *imagebuilder.BuildReport
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

		fmt.Println("Getting build report.")
		// Parse the build report from the ADO pipeline run logs.
		buildReport, err = imagebuilder.NewBuildReportFromLogs(logs)
		if err != nil {
			return fmt.Errorf("build in ADO failed, failed parsing build report from ADO pipeline run logs, err: %s", err)
		}

		o.logger.Debugw("Parsed build report from ADO logs", "buildReport", buildReport)
	} else {
		dryRunPipelineRunResult := pipelines.RunResult("Succeeded")
		pipelineRunResult = &dryRunPipelineRunResult
	}

	// TODO: Setting github outputs should happen outside buildInADO function.
	//  buildInADO should return required data and caller should handle it.
	// if run in github actions, set output parameters
	if o.ciSystem == GithubActions {
		fmt.Println("Setting GitHub outputs.")

		o.logger.Debugw("Extracted built images from ADO logs", "images", buildReport.Images, "architectures", buildReport.Architectures)

		imagesJSON, err := json.Marshal(buildReport.Images)
		if err != nil {
			return fmt.Errorf("cannot marshal list of images: %w", err)
		}

		architecturesJSON, err := json.Marshal(buildReport.Architectures)
		if err != nil {
			return fmt.Errorf("cannot marshal list of architectures: %w", err)
		}

		o.logger.Debugw("Set GitHub outputs", "images", string(imagesJSON), "architectures", string(architecturesJSON), "adoResult", string(*pipelineRunResult))

		err = actions.SetOutput("images", string(imagesJSON))
		if err != nil {
			return fmt.Errorf("cannot set images GitHub output: %w", err)
		}

		if err := actions.SetOutput("architectures", string(architecturesJSON)); err != nil {
			return fmt.Errorf("cannot set architectures GitHub output: %w", err)
		}

		err = actions.SetOutput("adoResult", string(*pipelineRunResult))
		if err != nil {
			return fmt.Errorf("cannot set adoResult GitHub output: %w", err)
		}
	}

	if o.buildReportPath != "" {
		err = imagebuilder.WriteReportToFile(buildReport, o.buildReportPath)
		if err != nil {
			return fmt.Errorf("failed writing build report to file: %w", err)
		}
	}

	// Handle the ADO pipeline run failure.
	if *pipelineRunResult == pipelines.RunResultValues.Failed || *pipelineRunResult == pipelines.RunResultValues.Unknown {
		return fmt.Errorf("build in ADO finished with status: %s", *pipelineRunResult)
	}
	return nil
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

func getTags(logger Logger, pr, sha string, templates []tags.Tag) ([]tags.Tag, error) {
	logger.Debugw("started building tags", "pr_number", pr, "commit_sha", sha, "templates", templates)

	logger.Debugw("building tagger options")
	var taggerOptions []tags.TagOption
	if len(pr) > 0 {
		taggerOptions = append(taggerOptions, tags.PRNumber(pr))
		logger.Debugw("pr number is set, adding tagger option", "pr_number", pr)
	}
	if len(sha) > 0 {
		taggerOptions = append(taggerOptions, tags.CommitSHA(sha))
		logger.Debugw("commit sha is set, adding tagger option", "commit_sha", sha)
	}

	taggerOptions = append(taggerOptions, tags.WithLogger(logger))
	logger.Debugw("added logger to tagger options")
	// build a tag from commit SHA
	tagger, err := tags.NewTagger(logger, templates, taggerOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed creating tagger instance: %w", err)
	}
	logger.Debugw("created tagger instance with options, starting parsing tags")
	p, err := tagger.ParseTags()
	if err != nil {
		return nil, fmt.Errorf("build tag: %w", err)
	}
	logger.Debugw("parsed tags successfully", "tags", p)

	return p, nil
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

	if o.adoPreviewRun && o.adoPreviewRunYamlPath == "" {
		errs = append(errs, fmt.Errorf("ado-preview-run-yaml-path flag is missing, please provide path to yaml file with ADO pipeline definition"))
	}

	if o.adoPreviewRunYamlPath != "" && !o.adoPreviewRun {
		errs = append(errs, fmt.Errorf("ado-preview-run-yaml-path flag is provided, but adoPreviewRun flag is not set to true"))
	}

	return errutil.NewAggregate(errs)
}

func (o *options) gatherOptions(flagSet *flag.FlagSet) *flag.FlagSet {
	flagSet.BoolVar(&o.silent, "silent", false, "Do not push build logs to stdout")
	flagSet.StringVar(&o.configPath, "config", "/config/image-builder-config.yaml", "Path to application config file")
	flagSet.StringVar(&o.context, "context", ".", "Path to build directory context")
	flagSet.StringVar(&o.name, "name", "", "name of the image to be built")
	flagSet.StringVar(&o.dockerfile, "dockerfile", "dockerfile", "Path to dockerfile file relative to context")
	flagSet.StringVar(&o.logDir, "log-dir", "/logs/artifacts", "Path to logs directory where GCB logs will be stored")
	flagSet.BoolVar(&o.debug, "debug", false, "Enable debug logging")
	flagSet.BoolVar(&o.dryRun, "dry-run", false, "Do not build the image, only print a ADO API call pipeline parameters")
	// TODO: What is expected value repo only or org/repo? How this flag influence an image builder behaviour?
	flagSet.StringVar(&o.orgRepo, "repo", "", "Load repository-specific configuration, for example, signing configuration")
	flagSet.Var(&o.tags, "tag", "Additional tag that the image will be tagged with. Optionally you can pass the name in the format name=value which will be used by export-tags")
	flagSet.StringVar(&o.tagsBase64, "tag-base64", "", "String representation of all tags encoded by base64. String representation must be in format as output of kyma-project/test-infra/pkg/tags.Tags.String() method")
	flagSet.Var(&o.buildArgs, "build-arg", "Flag to pass additional arguments to build dockerfile. It can be used in the name=value format.")
	flagSet.Var(&o.platforms, "platform", "Platform of the image that is built (default: linux/amd64,linux/arm64)")
	flagSet.BoolVar(&o.exportTags, "export-tags", false, "Export parsed tags as build-args into dockerfile. Each tag will have format TAG_x, where x is the tag name passed along with the tag")
	flagSet.BoolVar(&o.signOnly, "sign-only", false, "Only sign the image, do not build it")
	flagSet.Var(&o.imagesToSign, "images-to-sign", "Comma-separated list of images to sign. Only used when sign-only flag is set")
	flagSet.BoolVar(&o.adoPreviewRun, "ado-preview-run", false, "Trigger ADO pipeline in preview mode")
	flagSet.StringVar(&o.adoPreviewRunYamlPath, "ado-preview-run-yaml-path", "", "Path to yaml file with ADO pipeline definition to be used in preview mode")
	flagSet.BoolVar(&o.parseTagsOnly, "parse-tags-only", false, "Only parse tags and print them to stdout")
	flagSet.StringVar(&o.oidcToken, "oidc-token", "", "Token used to authenticate against Azure DevOps backend service")
	flagSet.StringVar(&o.azureAccessToken, "azure-access-token", "", "Token used to authenticate against Azure DevOps API")
	flagSet.StringVar(&o.tagsOutputFile, "tags-output-file", "/generated-tags.json", "Path to file where generated tags will be written as JSON")
	flagSet.BoolVar(&o.useGoInternalSAPModules, "use-go-internal-sap-modules", false, "Allow access to Go internal modules in ADO backend")
	flagSet.StringVar(&o.buildReportPath, "build-report-path", "", "Path to file where build report will be written as JSON")
	flagSet.BoolVar(&o.adoStateOutput, "ado-state-output", false, "Set output variables with result of image-buidler exececution")
	flagSet.StringVar(&o.target, "target", "", "Specify which build stage in the Dockerfile to use as the target")

	return flagSet
}

func main() {
	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	o := options{isCI: os.Getenv("CI") == "true"}
	o.gatherOptions(flagSet)
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		log.Fatalf("Failed to parse flags: %s", err)
	}

	var (
		zapLogger *zap.Logger
		err       error
	)
	if o.debug {
		zapLogger, err = zap.NewDevelopment()
	} else {
		zapLogger, err = zap.NewProduction()
	}
	if err != nil {
		log.Fatalf("Failed to initialize logger: %s", err)
	}
	o.logger = zapLogger.Sugar()

	// If running inside some CI system, determine which system is used
	if o.isCI {
		o.ciSystem, err = DetermineUsedCISystem()
		if err != nil {
			o.logger.Errorw("Failed to determine current ci system", "error", err)
			os.Exit(1)
		}

		o.gitState, err = LoadGitStateConfig(o.logger, o.ciSystem)
		if err != nil {
			o.logger.Errorw("Failed to load current git state", "error", err)
			os.Exit(1)
		}

		o.logger.Debugw("Git state loaded", "gitState", o.gitState)
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
			if o.adoStateOutput {
				adopipelines.SetVariable("signing_success", false, false, true)
			}

			fmt.Println(err)
			os.Exit(1)
		}

		if o.adoStateOutput {
			adopipelines.SetVariable("signing_success", true, false, true)
		}
		os.Exit(0)
	}

	if o.parseTagsOnly {
		logger := o.logger.With("command", "parse-tags-only")
		logger.Infow("Parsing tags")
		err = generateTags(logger, o)
		if err != nil {
			logger.Errorw("Parsing tags failed", "error", err)
			os.Exit(1)
		}
		logger.Infow("Tags parsed successfully")
		os.Exit(0)
	}
	err = buildInADO(o)
	if err != nil {
		o.logger.Errorw("Image build failed", "error", err, "JobType", o.gitState.JobType)
		os.Exit(1)
	}

	fmt.Println("Job's done.")
}

func generateTags(logger Logger, o options) error {
	logger.Infow("starting tag generation")
	logger.Debugw("getting the absolute path to the Dockerfile directory")
	// Get the absolute path to the dockerfile directory.
	dockerfileDirPath, err := getDockerfileDirPath(logger, o)
	if err != nil {
		return fmt.Errorf("failed to get dockerfile path: %w", err)
	}
	logger.Debugw("dockerfile directory path retrieved", "dockerfileDirPath", dockerfileDirPath)
	logger.Debugw("parsing tags from options")
	// Parse tags from the provided options.
	parsedTags, err := parseTags(logger, o)
	if err != nil {
		return fmt.Errorf("failed to parse tags from options: %w", err)
	}
	logger.Infow("tags parsed successfully", "parsedTags", parsedTags)
	logger.Debugw("converting parsed tags to JSON")
	jsonTags, err := tagsAsJSON(parsedTags)
	if err != nil {
		return fmt.Errorf("failed generating tags json representation: %w", err)
	}
	logger.Debugw("successfully generated image tags in JSON format", "tags", jsonTags)
	// Write tags to a file.
	if o.tagsOutputFile != "" {
		logger.Debugw("tags output file provided", "tagsOutputFile", o.tagsOutputFile)
		err = writeOutputFile(logger, o.tagsOutputFile, jsonTags)
		if err != nil {
			return fmt.Errorf("failed to write tags to file: %w", err)
		}
		logger.Infow("tags successfully written to file", "tagsOutputFile", o.tagsOutputFile, "generatedTags", jsonTags)
	}
	return nil
}

// writeOutputFile writes the provided data to the file specified by the path.
func writeOutputFile(logger Logger, path string, data []byte) error {
	logger.Debugw("writing generated tags to file", "tagsOutputFile", path)
	err := os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write tags to file: %s", err)
	}
	logger.Debugw("tags written to file")
	return nil
}

func tagsAsJSON(parsedTags []tags.Tag) ([]byte, error) {
	jsonTags, err := json.Marshal(parsedTags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tags to json, got error: %w", err)
	}
	return jsonTags, err
}

func parseTags(logger Logger, o options) ([]tags.Tag, error) {
	logger.Debugw("starting to parse tags")
	var (
		pr  string
		sha string
	)

	logger.Debugw("reading git state for event type")
	if !o.gitState.isPullRequest && o.gitState.BaseCommitSHA != "" {
		sha = o.gitState.BaseCommitSHA
		logger.Debugw("running for push event, base commit SHA found", "sha", sha)
	}
	if o.gitState.isPullRequest && o.gitState.PullRequestNumber > 0 {
		pr = fmt.Sprint(o.gitState.PullRequestNumber)
		logger.Debugw("Running for pull request event, PR number found", "pr-number", pr)
	}

	// TODO (dekiel): Tags provided as base64 encoded string should be parsed and added to the tags list when parsing flags.
	//   This way all tags are available in the tags list from thr very beginning of execution and can be used in any process.

	logger.Debugw("checking if base64 encoded tags are provided")
	// read tags from base64 encoded string if provided
	if o.tagsBase64 != "" {
		logger.Debugw("base64 encoded tags provided, starting to decode", "tagsBase64", o.tagsBase64)
		decoded, err := base64.StdEncoding.DecodeString(o.tagsBase64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 encoded tags, error: %w", err)
		}
		logger.Debugw("tags successfully decoded", "decoded", string(decoded))
		splitedTags := strings.Split(string(decoded), ",")
		logger.Debugw("splitted decoded tags", "splitedTags", splitedTags)
		for _, tag := range splitedTags {
			logger.Debugw("adding tag", "tag", tag)
			err = o.tags.Set(tag)
			if err != nil {
				return nil, fmt.Errorf("failed to set tag, tag: %s, error: %w", tag, err)
			}
			logger.Debugw("tag set successfully")
		}
		logger.Debugw("all base64 encoded tags successfully added", "tags", o.tags.String())
	} else {
		logger.Debugw("no base64 encoded tags provided")
	}

	logger.Debugw("getting default tag")
	defaultTag, err := getDefaultTag(logger, o)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve default tag, error: %w", err)
	}
	logger.Debugw("default tag retrieved", "defaultTag", defaultTag)

	logger.Debugw("parsing tags")
	parsedTags, err := getTags(logger, pr, sha, append(o.tags, defaultTag))
	if err != nil {
		return nil, fmt.Errorf("failed to parse tags: %w", err)
	}
	logger.Debugw("tags parsed successfully", "parsedTags", parsedTags)

	return parsedTags, nil
}

// getDefaultTag returns the default tag based on the read git state.
// The function provide default tag for pull request or commit.
// The default tag is read from the provided 'options' struct.
func getDefaultTag(logger Logger, o options) (tags.Tag, error) {
	logger.Debugw("reading gitstate data")
	if o.gitState.isPullRequest && o.gitState.PullRequestNumber > 0 {
		logger.Debugw("pull request number provided, returning default pr tag")
		return o.DefaultPRTag, nil
	}
	if len(o.gitState.BaseCommitSHA) > 0 {
		o.logger.Debugw("commit sha provided, returning default commit tag")
		return o.DefaultCommitTag, nil
	}
	return tags.Tag{}, fmt.Errorf("could not determine default tag, no pr number or commit sha provided")
}

func getDockerfileDirPath(logger Logger, o options) (string, error) {
	logger.Debugw("starting to get Dockerfile directory path", "dockerfile", o.dockerfile, "context", o.context)
	// Get the absolute path to the build context directory.
	context, err := filepath.Abs(o.context)
	if err != nil {
		return "", fmt.Errorf("could not get absolute path to build context directory: %w", err)
	}
	logger.Debugw("successfully retrieved absolute path to context directory", "absolute_path", context)
	// Get the absolute path to the dockerfile.
	dockerfileDirPath := filepath.Join(context, filepath.Dir(o.dockerfile))
	logger.Debugw("dockerfile directory path constructed", "dockerfileDirPath", dockerfileDirPath)
	return dockerfileDirPath, err
}
