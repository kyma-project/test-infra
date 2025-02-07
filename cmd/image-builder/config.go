package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strconv"

	"github.com/google/go-github/v48/github"
	adoPipelines "github.com/kyma-project/test-infra/pkg/azuredevops/pipelines"
	"github.com/kyma-project/test-infra/pkg/sign"
	"github.com/kyma-project/test-infra/pkg/tags"
	"gopkg.in/yaml.v3"
)

type CISystem string

// Enum of supported CI/CD systems to read data from
const (
	Prow          CISystem = "Prow"
	GithubActions CISystem = "GithubActions"
	AzureDevOps   CISystem = "AzureDevOps"
	Jenkins       CISystem = "Jenkins"
)

type Config struct {
	AdoConfig adoPipelines.Config `yaml:"ado-config,omitempty" json:"ado-config,omitempty"`
	// Registry is URL where clean build should land.
	Registry Registry `yaml:"registry" json:"registry"`
	// DevRegistry is Registry URL where development/dirty images should land.
	// If not set then the Registry field is used.
	// This field is only valid when running in CI (CI env variable is set to `true`)
	DevRegistry Registry `yaml:"dev-registry" json:"dev-registry"`
	// Cache options that are directly related to kaniko flags
	Cache CacheConfig `yaml:"cache" json:"cache"`
	// Default Tag template used for images build on commit.
	// The value can be a go-template string or literal tag value string.
	// See tags.Tag struct for more information and available fields
	DefaultCommitTag tags.Tag `yaml:"default-commit-tag" json:"default-commit-tag"`
	// Default Tag template used for images build on pull request.
	// The value can be a go-template string or literal tag value string.
	// See tags.Tag struct for more information and available fields
	DefaultPRTag tags.Tag `yaml:"default-pr-tag" json:"default-pr-tag"`
	// LogFormat defines the format kaniko logs are projected.
	// Supported formats are 'color', 'text' and 'json'. Default: 'color'
	LogFormat string `yaml:"log-format" json:"log-format"`
	// Set this option to strip timestamps out of the built image and make it Reproducible.
	Reproducible bool `yaml:"reproducible" json:"reproducible"`
	// SignConfig contains custom configuration of signers
	// as well as org/repo mapping of enabled signers in specific repository
	SignConfig SignConfig `yaml:"sign-config" json:"sign-config"`
}

type SignConfig struct {
	// EnabledSigners contains org/repo mapping of enabled signers for each repository
	// Use * to enable signer for all repositories
	EnabledSigners map[string][]string `yaml:"enabled-signers" json:"enabled-signers"`
	// Signers contains configuration for multiple signing backends, which can be used to sign resulting image
	Signers []sign.SignerConfig `yaml:"signers" json:"signers"`
}

type CacheConfig struct {
	// Enabled sets if kaniko cache is enabled or not
	Enabled bool `yaml:"enabled" json:"enabled"`
	// CacheRunLayers sets if kaniko should cache run layers
	CacheRunLayers bool `yaml:"cache-run-layers" json:"cache-run-layers"`
	// CacheCopyLayers sets if kaniko should cache copy layers
	CacheCopyLayers bool `yaml:"cache-copy-layers" json:"cache-copy-layers"`
	// Remote Docker directory used for cache
	CacheRepo string `yaml:"cache-repo" json:"cache-repo"`
}

// ParseConfig parses yaml configuration into Config
func (c *Config) ParseConfig(f []byte) error {
	return yaml.Unmarshal(f, c)
}

type Variants map[string]map[string]string

// GetVariants fetches variants from provided file.
// If variant flag is used, it fetches the requested variant.
func GetVariants(variant string, f string, fileGetter func(string) ([]byte, error)) (Variants, error) {
	var v Variants
	b, err := fileGetter(f)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		// variant file not found, skipping
		return nil, nil
	}
	if err := yaml.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	if variant != "" {
		va, ok := v[variant]
		if !ok {
			return nil, fmt.Errorf("requested variant '%s', but it's not present in variants.yaml file", variant)
		}
		return Variants{variant: va}, nil
	}
	return v, nil
}

// Registry is a custom type that defines a destination registry provided by config.yaml
type Registry []string

// UnmarshalYAML provides functionality to unmarshal Registry field if it's a string or a list.
// This functionality ensures, that both use cases are supported and there are no breaking changes in the config
func (r *Registry) UnmarshalYAML(value *yaml.Node) error {
	var reg string
	if err := value.Decode(&reg); err == nil {
		*r = append(*r, reg)
		return nil
	}
	var regs []string
	if err := value.Decode(&regs); err != nil {
		return err
	}
	*r = regs
	return nil
}

// GitStateConfig holds information about repository and specific commit
// from which image should be build.
// It also contains information whether job is presubmit or postsubmit
type GitStateConfig struct {
	// Name of the source repository
	RepositoryName string
	// Name of the source repository's owner
	RepositoryOwner string
	// Type of the job, allowed values "presubmit" or "postsubmit"
	JobType string
	// Number of the pull request for presubmit job
	PullRequestNumber int
	// Commit SHA for base branch or tag
	BaseCommitSHA string
	// Base branch or tag
	BaseCommitRef string
	// Commit SHA for head of the pull request
	PullHeadCommitSHA string
	// isPullRequest contains information whether event which triggered the job was from pull request
	isPullRequest bool
}

func (gitState GitStateConfig) IsPullRequest() bool {
	return gitState.isPullRequest
}

// TODO (dekiel): Add logger parameter to all functions reading a git state.
func LoadGitStateConfig(logger Logger, ciSystem CISystem) (GitStateConfig, error) {
	switch ciSystem {
	// Load from env specific for Azure DevOps and Prow Jobs
	case AzureDevOps, Prow:
		return loadADOGitState()
	// Load from env specific for Github Actions
	case GithubActions:
		return loadGithubActionsGitState()
	case Jenkins:
		return loadJenkinsGitState(logger)
	default:
		// Unknown CI System, return error and empty git state
		return GitStateConfig{}, fmt.Errorf("unknown ci system, got %s", ciSystem)
	}
}

func loadADOGitState() (GitStateConfig, error) {
	var pullNumber int

	repoName, present := os.LookupEnv("REPO_NAME")
	if !present {
		return GitStateConfig{}, fmt.Errorf("REPO_NAME environment variable is not set, please set it to valid repository name")
	}

	repoOwner, present := os.LookupEnv("REPO_OWNER")
	if !present {
		return GitStateConfig{}, fmt.Errorf("REPO_OWNER environment variable is not set, please set it to valid repository owner")
	}

	jobType, present := os.LookupEnv("JOB_TYPE")
	if !present {
		return GitStateConfig{}, fmt.Errorf("JOB_TYPE environment variable is not set, please set it to valid job type")
	}
	if !slices.Contains(adoPipelines.GetValidJobTypes(), jobType) {
		return GitStateConfig{}, fmt.Errorf("image builder is running for unsupported event %s", jobType)
	}

	pullNumberString, isPullNumberSet := os.LookupEnv("PULL_NUMBER")
	if jobType == "presubmit" {
		if !isPullNumberSet {
			return GitStateConfig{}, fmt.Errorf("PULL_NUMBER environment variable is not set, please set it to valid pull request number")
		}

		pullRequest, err := strconv.Atoi(pullNumberString)
		if err != nil {
			return GitStateConfig{}, fmt.Errorf("PULL_NUMBER environment variable contains invalid value, please set it to correct integer PR number: %w", err)
		}
		pullNumber = pullRequest
	}

	baseSHA, present := os.LookupEnv("PULL_BASE_SHA")
	if !present {
		return GitStateConfig{}, fmt.Errorf("PULL_BASE_SHA environment variable is not set, please set it to valid pull base SHA")
	}

	pullSHA, present := os.LookupEnv("PULL_PULL_SHA")
	if !present && jobType == "presubmit" {
		return GitStateConfig{}, fmt.Errorf("PULL_PULL_SHA environment variable is not set, please set it to valid pull head SHA")
	}

	return GitStateConfig{
		RepositoryName:    repoName,
		RepositoryOwner:   repoOwner,
		JobType:           jobType,
		PullRequestNumber: pullNumber,
		BaseCommitSHA:     baseSHA,
		PullHeadCommitSHA: pullSHA,
		isPullRequest:     pullNumber > 0 && pullSHA != "",
	}, nil
}

func loadGithubActionsGitState() (GitStateConfig, error) {
	eventName, present := os.LookupEnv("GITHUB_EVENT_NAME")
	if !present {
		return GitStateConfig{}, fmt.Errorf("the GITHUB_EVENT_NAME environment variable is not set.  Please ensure the image-builder is running in GitHub environment")
	}
	eventPayloadPath, present := os.LookupEnv("GITHUB_EVENT_PATH")
	if !present {
		return GitStateConfig{}, fmt.Errorf("the GITHUB_EVENT_PATH environment variable is not set. Please ensure the image-builder is running in GitHub environment")
	}
	// For PR and push events commit sha will be fetched from event payload
	commitSHA, present := os.LookupEnv("GITHUB_SHA")
	if !present && (eventName != "pull_request_target" && eventName != "push") {
		return GitStateConfig{}, fmt.Errorf("the GITHUB_SHA environment variable is not set, it should be set to HEAD commit SHA. Please ensure the image-builder is running in GitHub environment")
	}
	// For PR and push events commit ref will be fetched from event payload
	gitRef, present := os.LookupEnv("GITHUB_REF")
	if !present && (eventName != "pull_request_target" && eventName != "push") {
		return GitStateConfig{}, fmt.Errorf("the GITHUB_REF environment variable is not set, it should be set to current ref. Please ensure the image-builder is running in GitHub environment")
	}

	// Read event payload file from runner
	data, err := os.ReadFile(eventPayloadPath)
	if err != nil {
		return GitStateConfig{}, fmt.Errorf("failed to read content of event payload file: %w", err)
	}

	// Handle different events types
	switch eventName {
	case "pull_request_target":
		var payload github.PullRequestEvent
		err = json.Unmarshal(data, &payload)
		if err != nil {
			return GitStateConfig{}, fmt.Errorf("failed to parse event payload: %s", err)
		}

		return GitStateConfig{
			RepositoryName:    *payload.Repo.Name,
			RepositoryOwner:   *payload.Repo.Owner.Login,
			JobType:           "presubmit",
			PullRequestNumber: *payload.Number,
			BaseCommitSHA:     *payload.PullRequest.Base.SHA,
			PullHeadCommitSHA: *payload.PullRequest.Head.SHA,
			isPullRequest:     true,
		}, nil

	case "push":
		var payload github.PushEvent
		err = json.Unmarshal(data, &payload)
		if err != nil {
			return GitStateConfig{}, fmt.Errorf("failed to parse event payload: %s", err)
		}
		return GitStateConfig{
			RepositoryName:  *payload.Repo.Name,
			RepositoryOwner: *payload.Repo.Owner.Login,
			JobType:         "postsubmit",
			BaseCommitSHA:   *payload.HeadCommit.ID,
		}, nil

	case "workflow_dispatch":
		var payload github.WorkflowDispatchEvent
		err = json.Unmarshal(data, &payload)
		if err != nil {
			return GitStateConfig{}, fmt.Errorf("failed to parse event payload: %s", err)
		}
		return GitStateConfig{
			RepositoryName:  *payload.Repo.Name,
			RepositoryOwner: *payload.Repo.Owner.Login,
			JobType:         "workflow_dispatch",
			BaseCommitSHA:   commitSHA,
			BaseCommitRef:   gitRef,
		}, nil

	case "schedule":
		// There is nostruct for schedule event in github package
		var payload struct {
			Repo github.Repository `json:"repository,omitempty"`
		}
		err = json.Unmarshal(data, &payload)
		if err != nil {
			return GitStateConfig{}, fmt.Errorf("failed to parse event payload: %s", err)
		}
		return GitStateConfig{
			RepositoryName:  *payload.Repo.Name,
			RepositoryOwner: *payload.Repo.Owner.Login,
			JobType:         "schedule",
			BaseCommitSHA:   commitSHA,
			BaseCommitRef:   gitRef,
		}, nil

	case "merge_group":
		var payload github.MergeGroupEvent
		err = json.Unmarshal(data, &payload)
		if err != nil {
			return GitStateConfig{}, fmt.Errorf("failed to parse event payload: %s", err)
		}
		return GitStateConfig{
			RepositoryName:    *payload.Repo.Name,
			RepositoryOwner:   *payload.Repo.Owner.Login,
			JobType:           "merge_group",
			BaseCommitSHA:     commitSHA,
			BaseCommitRef:     gitRef,
			PullHeadCommitSHA: *payload.MergeGroup.HeadSHA,
			isPullRequest:     true,
		}, nil

	default:
		return GitStateConfig{}, fmt.Errorf("GITHUB_EVENT_NAME environment variable is set to unsupported value \"%s\", please set it to supported value", eventName)
	}
}

// loadJenkinsGitState loads git state from environment variables specific for Jenkins.
func loadJenkinsGitState(logger Logger) (GitStateConfig, error) {
	// Load from env specific for Jenkins Jobs
	prID, isPullRequest := os.LookupEnv("CHANGE_ID")
	gitURL, present := os.LookupEnv("GIT_URL")
	if !present {
		return GitStateConfig{}, fmt.Errorf("GIT_URL environment variable is not set, please set it to valid git URL")
	}

	owner, repo, err := extractOwnerAndRepoFromGitURL(logger, gitURL)
	if err != nil {
		return GitStateConfig{}, fmt.Errorf("failed to extract owner and repository from git URL %s: %w", gitURL, err)
	}

	gitState := GitStateConfig{
		RepositoryName:  repo,
		RepositoryOwner: owner,
	}

	if isPullRequest {
		pullNumber, err := strconv.Atoi(prID)
		if err != nil {
			return GitStateConfig{}, fmt.Errorf("failed to convert prID string variable to integer: %w", err)
		}

		baseRef, present := os.LookupEnv("CHANGE_BRANCH")
		if !present {
			return GitStateConfig{}, fmt.Errorf("CHANGE_BRANCH environment variable is not set, please set it to valid base branch name")
		}

		// In Jenkins, the GIT_COMMIT is head commit SHA for pull request
		// See: https://github.tools.sap/kyma/oci-image-builder/issues/165
		headCommitSHA, present := os.LookupEnv("GIT_COMMIT")
		if !present {
			return GitStateConfig{}, fmt.Errorf("GIT_COMMIT environment variable is not set, please set it to valid head commit SHA")
		}

		baseCommitSHA, present := os.LookupEnv("CHANGE_BASE_SHA")
		if !present {
			return GitStateConfig{}, fmt.Errorf("CHANGE_BASE_SHA environment variable is not set, please set it to valid base commit SHA")
		}

		gitState.JobType = "presubmit"
		gitState.PullRequestNumber = pullNumber
		gitState.BaseCommitSHA = baseCommitSHA
		gitState.PullHeadCommitSHA = headCommitSHA
		gitState.BaseCommitRef = baseRef
		gitState.isPullRequest = true

		return gitState, nil
	}

	baseCommitSHA, present := os.LookupEnv("GIT_COMMIT")
	if !present {
		return GitStateConfig{}, fmt.Errorf("GIT_COMMIT environment variable is not set, please set it to valid commit SHA")
	}

	gitState.JobType = "postsubmit"
	gitState.BaseCommitSHA = baseCommitSHA

	return gitState, nil
}

func extractOwnerAndRepoFromGitURL(logger Logger, gitURL string) (string, string, error) {
	re := regexp.MustCompile(`.*/(?P<owner>.*)/(?P<repo>.*).git`)
	matches := re.FindStringSubmatch(gitURL)

	logger.Debugw("Extracted matches from git URL", "matches", matches, "gitURL", gitURL)

	if len(matches) != 3 {
		return "", "", fmt.Errorf("failed to extract owner and repository from git URL")
	}

	owner := matches[re.SubexpIndex("owner")]
	repo := matches[re.SubexpIndex("repo")]

	logger.Debugw("Extracted owner from git URL", "owner", owner)
	logger.Debugw("Extracted repository from git URL", "repo", repo)

	return owner, repo, nil
}

// DetermineUsedCISystem return CISystem bind to system in which image builder is running or error if unknown
// It is used to avoid getting env variables in multiple parts of image builder
func DetermineUsedCISystem() (CISystem, error) {
	// Use system functions in production implementation
	return determineUsedCISystem(os.Getenv, os.LookupEnv)
}

// Additional private function for testing purposes.
// It allows us to mock os.Getenv and os.LookupEnv during tests, keeping logic valid
// Reason to introduce that is lack of possibility to override variables in CI systems
func determineUsedCISystem(envGetter func(key string) string, envLookup func(key string) (string, bool)) (CISystem, error) {
	// GITHUB_ACTIONS environment variable is always set to true in github actions workflow
	// See: https://docs.github.com/en/actions/learn-github-actions/variables#default-environment-variables
	isGithubActions := envGetter("GITHUB_ACTIONS")
	if isGithubActions == "true" {
		return GithubActions, nil
	}

	// PROW_JOB_ID environment variables contains ID of prow job
	// See: https://docs.prow.k8s.io/docs/jobs/#job-environment-variables
	_, isProwJob := envLookup("PROW_JOB_ID")
	if isProwJob {
		return Prow, nil
	}

	// BUILD_BUILDID environment variable is set in Azure DevOps pipeline
	// See: https://learn.microsoft.com/en-us/azure/devops/pipelines/build/variables?view=azure-devops&tabs=yaml#build-variables-devops-services
	_, isAdo := envLookup("BUILD_BUILDID")
	if isAdo {
		return AzureDevOps, nil
	}

	// JENKINS_HOME environment variable is set in Jenkins
	_, isJenkins := envLookup("JENKINS_HOME")
	if isJenkins {
		return Jenkins, nil
	}

	return "", fmt.Errorf("cannot determine ci system: unknown system")
}
