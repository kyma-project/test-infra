package client

import (
	"fmt"

	"github.com/kyma-project/test-infra/development/gcp/pkg/logging"
	prowflagutil "k8s.io/test-infra/prow/flagutil"
	"k8s.io/test-infra/prow/git/v2"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/repoowners"
)

type GithubClientConfig struct {
	prowflagutil.GitHubOptions
	DryRun bool
}

type GithubClient struct {
	github.Client
}

type ClientsAgent struct {
	//	GithubConfig prowflagutil.GitHubOptions
	//	GitConfig    git.ClientFactoryOpts
	//	OwnersConfig OwnersConfig
	//	DryRun       bool
	logging.LoggerInterface
	GithubClient github.Client
	GitClient    git.ClientFactory
	OwnersClient repoowners.Client
}

type AgentOption func(*ClientsAgent) error
type GithubClientOption func(*GithubClientConfig) error

// TODO: With github, git, woners client
func NewClientsAgent(options ...AgentOption) (*ClientsAgent, error) {
	ca := &ClientsAgent{
		// GithubConfig: prowflagutil.GitHubOptions{},
		// GitConfig:    git.ClientFactoryOpts{},
		// OwnersConfig: OwnersConfig{},
		// DryRun:       false,
		GithubClient: nil,
		GitClient:    nil,
		OwnersClient: repoowners.Client{},
	}

	for _, opt := range options {
		err := opt(ca)
		if err != nil {
			return nil, fmt.Errorf("failed applying functional option: %w", err)
		}
	}

	return ca, nil
}

func WithLogger(logger *logging.LoggerInterface) AgentOption {
	return func(ca *ClientsAgent) error {
		ca.LoggerInterface = *logger
		return nil
	}
}

func NewGithubClient(options ...GithubClientOption) (*GithubClient, error) {
	conf := &GithubClientConfig{
		GitHubOptions: prowflagutil.GitHubOptions{},
		DryRun:        false,
	}

	for _, opt := range options {
		err := opt(conf)
		if err != nil {
			return nil, fmt.Errorf("failed applying functional option: %w", err)
		}
	}
	client, err := conf.GitHubOptions.GitHubClient(conf.DryRun)
	if err != nil {
		return nil, err
	}
	return &GithubClient{
		Client: client,
	}, nil
}

func WithTokenPath(tokenpath string) GithubClientOption {
	return func(conf *GithubClientConfig) error {
		conf.TokenPath = tokenpath
		return nil
	}
}
