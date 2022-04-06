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
	// ClonedRepos is a map with information about already cloned repositories.
	// A map keys represent hold org/repo and values are a path to a repository root.
	clonedRepos map[string]string
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

	// Run provided GitClientOption configuration options.
	for _, opt := range options {
		err := opt(o)
		if err != nil {
			return nil, fmt.Errorf("failed applying functional option: %w", err)
		}
	}

	// Check mandatory option is provided.
	if o.githubClient == nil {
		return nil, fmt.Errorf("github client not provided")
	}

	tokenGenerator := secret.GetTokenGenerator(o.tokenPath)

	gitFactory, err := o.GitClient(o.githubClient, tokenGenerator, secret.Censor, false)
	if err != nil {
		return nil, err
	}
	gitClient := &GitClient{}
	// Initialize map to enable writing to it in methods.
	gitClient.clonedRepos = make(map[string]string)
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

// WithTokenPath is a client constructor configuration option passing git token file path.
func WithTokenPath(tokenPath string) GitClientOption {
	return func(conf *GitClientConfig) error {
		conf.tokenPath = tokenPath
		return nil
	}
}

// GetGitRepoClient provide instance of git repository client. It will clone repository on first use.
// If repository was already cloned, a new repository client will be created from local repository.
// During creation from local repository a fetch from upstream is executed.
func (c *GitClient) GetGitRepoClient(org, repo string) (git.RepoClient, string, error) {
	// Check if repository was already cloned and reuse it.
	if path, ok := c.clonedRepos[fmt.Sprintf("%s/%s", org, repo)]; ok {
		// Create repository client from already cloned local repository.
		gitRepoClient, err := c.ClientFromDir(org, repo, path)
		if err != nil {
			return nil, "", fmt.Errorf("failed create git repository client from directory, org: %s, repo: %s, directory: %s, error: %w", org, repo, path, err)
		}
		// Fetch changes from upstream.
		err = gitRepoClient.Fetch()
		if err != nil {
			return nil, "", fmt.Errorf("failed fetch repostiory, org: %s, repo: %s, error: %w", org, repo, err)
		}
		return gitRepoClient, path, nil
	} else {
		// Create repository client for new repository by cloning from github.
		gitRepoClient, err := c.ClientFor(org, repo)
		if err != nil {
			return nil, "", fmt.Errorf("failed create git repository client, org: %s, repo: %s, error: %w", org, repo, err)
		}
		// Save repository local path to reuse it for creation new repository clients.
		c.clonedRepos[fmt.Sprintf("%s/%s", org, repo)] = gitRepoClient.Directory()
		return gitRepoClient, c.clonedRepos[fmt.Sprintf("%s/%s", org, repo)], nil
	}
}
