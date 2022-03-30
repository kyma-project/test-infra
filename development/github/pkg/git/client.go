package git

import (
	"fmt"

	"github.com/kyma-project/test-infra/development/github/pkg/client/v2"
	"k8s.io/test-infra/prow/git/v2"
)

// GitClient is a client to interact with git.
// It's build on top of k8s git.ClientFactory.
type GitClient struct {
	git.ClientFactory
}

// GitClientConfig holds configuration for GitClient.
type GitClientConfig struct {
	git.ClientFactoryOpts
	githubClient *client.GithubClient
}

// GitClientOption is a client constructor configuration option passing configuration to the client constructor.
type GitClientOption func(*GitClientConfig) error

// NewGitClient is a constructor function for GitClient.
// A constructed client can be configured by providing GitClientOptions.
func NewGitClient(options ...GitClientOption) (*GitClient, error) {
	var gitClient *GitClient

	conf := &GitClientConfig{}

	for _, opt := range options {
		err := opt(conf)
		if err != nil {
			return nil, fmt.Errorf("failed applying functional option: %w", err)
		}
	}
	if conf.githubClient == nil {
		return nil, fmt.Errorf("github client not provided")
	}
	gitUser := func() (name, email string, err error) {
		user, err := conf.githubClient.BotUser()
		if err != nil {
			return "", "", err
		}
		name = user.Name
		email = user.Email
		return name, email, nil
	}
	opts := git.ClientFactoryOpts{
		GitUser: gitUser,
	}
	factory, err := git.NewClientFactory(opts.Apply)
	if err != nil {
		return nil, err
	}
	gitClient.ClientFactory = factory
	return gitClient, nil
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
