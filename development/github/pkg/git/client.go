package git

import (
	"fmt"

	"github.com/kyma-project/test-infra/development/github/pkg/client/v2"
	"k8s.io/test-infra/prow/config/secret"
	"k8s.io/test-infra/prow/flagutil"
	"k8s.io/test-infra/prow/git/v2"
)

// GitClient is a client to interact with git.
// It's build on top of k8s git.ClientFactory.
type GitClient struct {
	git.ClientFactory
}

// GitClientConfig holds configuration for GitClient.
type GitClientConfig struct {
	flagutil.GitOptions
	tokenPath string
	// git.ClientFactoryOpts
	githubClient *client.GithubClient
}

// GitClientOption is a client constructor configuration option passing configuration to the client constructor.
type GitClientOption func(*GitClientConfig) error

// NewGitClient is a constructor function for wrapper of GitClient.
// A constructed client can be configured by providing GitClientOptions.
func (o *GitClientConfig) NewGitClient(options ...GitClientOption) (*GitClient, error) {
	// var gitClient *GitClient

	for _, opt := range options {
		err := opt(o)
		if err != nil {
			return nil, fmt.Errorf("failed applying functional option: %w", err)
		}
	}
	if o.githubClient == nil {
		return nil, fmt.Errorf("github client not provided")
	}

	tokenGenerator := secret.GetTokenGenerator(o.tokenPath)

	gitFactory, err := o.GitClient(o.githubClient, tokenGenerator, secret.Censor, false)
	if err != nil {
		return nil, err
	}
	gitClient := &GitClient{}
	gitClient.ClientFactory = gitFactory
	return gitClient, err
}

// WithGithubClient is a client constructor configuration option passing GithubClient instance.
func WithGithubClient(githubClient *client.GithubClient) GitClientOption {
	return func(conf *GitClientConfig) error {
		if conf.githubClient == nil {
			conf.githubClient = githubClient
			return nil
		} else {
			return fmt.Errorf("github client already defined")
		}
	}
}

// WithGithubClient is a client constructor configuration option passing GithubClient instance.
func WithTokenPath(tokenPath string) GitClientOption {
	return func(conf *GitClientConfig) error {
		conf.tokenPath = tokenPath
		return nil
	}
}
