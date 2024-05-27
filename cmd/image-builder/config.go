package main

import (
	"encoding/json"
	"fmt"
	"os"
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
	// TagTemplate is go-template field that defines the format of the $_TAG substitution.
	// See tags.Tag struct for more information and available fields
	TagTemplate tags.Tag `yaml:"tag-template" json:"tag-template"`
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
	// Commit SHA for base branch
	BaseCommitSHA string
	// Commit SHA for head of the pull request
	PullHeadCommitSHA string
	// isPullRequest contains information whether event which triggered the job was from pull request
	isPullRequest bool
}

func (gitState GitStateConfig) IsPullRequest() bool {
	return gitState.isPullRequest
}

func LoadGitStateConfig(ciSystem CISystem) (GitStateConfig, error) {
	switch ciSystem {
	// Load from env specific for Azure DevOps and Prow Jobs
	case AzureDevOps, Prow:
		return loadADOGitState()
	// Load from env specific for Github Actions
	case GithubActions:
		return loadGithubActionsGitState()
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
	if !slices.Contains([]string{"presubmit", "postsubmit"}, jobType) {
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
		return GitStateConfig{}, fmt.Errorf("GITHUB_EVENT_NAME environment variable is not set, please set it to valid event name")
	}
	eventPayloadPath, present := os.LookupEnv("GITHUB_EVENT_PATH")
	if !present {
		return GitStateConfig{}, fmt.Errorf("GITHUB_EVENT_PATH environment variable is not set, please set it to valid path to event file")
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
	default:
		return GitStateConfig{}, fmt.Errorf("GITHUB_EVENT_NAME environment variable is set to unsupported value \"%s\", please set it to supported value", eventName)
	}
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

	isAdo := envGetter("CI_SYSTEM") == "AzureDevOps"
	if isAdo {
		return AzureDevOps, nil
	}

	return "", fmt.Errorf("cannot determine ci system: unknown system")
}
