package git

import (
	"fmt"

	"sigs.k8s.io/prow/prow/config/secret"
	"sigs.k8s.io/prow/prow/flagutil"
	"sigs.k8s.io/prow/prow/git/v2"
	"sigs.k8s.io/prow/prow/github"
)

type Client interface {
	git.ClientFactory
	RepoClient
}

type RepoClient interface {
	GetGitRepoClient(org, repo string) (git.RepoClient, string, error)
	GetGitRepoClientFromDir(org, repo, dir string) (git.RepoClient, string, error)
}

// Client is a client to interact with git.
// It's build on top of k8s git.ClientFactory.
type client struct {
	git.ClientFactory
	// ClonedRepos is a map with information about already cloned repositories.
	// A map keys represent hold org/repo and values are a path to a repository root.
	clonedRepos map[string]string
}

// ClientConfig holds configuration for Client.
type ClientConfig struct {
	flagutil.GitOptions
	// tokenPath is a path to the file with GitHub user personal token.
	tokenPath        string
	githubUserClient github.UserClient
}

// ClientOption is a client constructor configuration option passing configuration to the client constructor.
type ClientOption func(*ClientConfig) error

// NewClient is a constructor function for wrapper of Client.
// A constructed client can be configured by providing ClientOptions.
func (o *ClientConfig) NewClient(options ...ClientOption) (Client, error) {

	// Run provided ClientOption configuration options.
	for _, opt := range options {
		err := opt(o)
		if err != nil {
			return nil, fmt.Errorf("failed applying functional option: %w", err)
		}
	}

	// Check mandatory option is provided.
	if o.githubUserClient == nil {
		return nil, fmt.Errorf("github client not provided")
	}

	tokenGenerator := secret.GetTokenGenerator(o.tokenPath)

	gitFactory, err := o.GitClient(o.githubUserClient, tokenGenerator, secret.Censor, false)
	if err != nil {
		return nil, err
	}
	gitClient := &client{}
	// Initialize map to enable writing to it in methods.
	gitClient.clonedRepos = make(map[string]string)
	gitClient.ClientFactory = gitFactory
	return gitClient, err
}

// WithGithubClient is a client constructor configuration option passing GithubClient instance.
func WithGithubClient(githubClient github.UserClient) ClientOption {
	return func(conf *ClientConfig) error {
		if conf.githubUserClient == nil {
			conf.githubUserClient = githubClient
			return nil
		}
		return fmt.Errorf("github client already defined")

	}
}

// WithTokenPath is a client constructor configuration option passing git token file path.
func WithTokenPath(tokenPath string) ClientOption {
	return func(conf *ClientConfig) error {
		conf.tokenPath = tokenPath
		return nil
	}
}

// GetGitRepoClient provide instance of git repository client. It will clone repository on first use.
// If repository was already cloned, a new repository client will be created from local repository.
// During creation from local repository a fetch from upstream is executed.
func (c *client) GetGitRepoClient(org, repo string) (git.RepoClient, string, error) {
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
			return nil, "", fmt.Errorf("failed fetch repository, org: %s, repo: %s, error: %w", org, repo, err)
		}
		return gitRepoClient, path, nil
	}
	// Create repository client for new repository by cloning from github.
	gitRepoClient, err := c.ClientFor(org, repo)
	if err != nil {
		return nil, "", fmt.Errorf("failed create git repository client, org: %s, repo: %s, error: %w", org, repo, err)
	}
	// Save repository local path to reuse it for creation new repository clients.
	c.clonedRepos[fmt.Sprintf("%s/%s", org, repo)] = gitRepoClient.Directory()
	return gitRepoClient, c.clonedRepos[fmt.Sprintf("%s/%s", org, repo)], nil
}

// GetGitRepoClientFromDir provide instance of git repository client. It will clone repository on first use.
// If repository was already cloned, a new repository client will be created from local repository.
// During creation from local repository a fetch from upstream is executed.
func (c *client) GetGitRepoClientFromDir(org, repo, dir string) (git.RepoClient, string, error) {
	// Create repository client from already cloned local repository.
	gitRepoClient, err := c.ClientFromDir(org, repo, dir)
	if err != nil {
		return nil, "", fmt.Errorf("failed create git repository client from directory, org: %s, repo: %s, directory: %s, error: %w", org, repo, dir, err)
	}
	// Fetch changes from upstream.
	err = gitRepoClient.Fetch()
	if err != nil {
		return nil, "", fmt.Errorf("failed fetch repository, org: %s, repo: %s, error: %w", org, repo, err)
	}
	// Save repository local path to reuse it for creation new repository clients.
	c.clonedRepos[fmt.Sprintf("%s/%s", org, repo)] = dir
	return gitRepoClient, dir, nil
}
