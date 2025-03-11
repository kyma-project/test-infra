package git

import (
	"fmt"

	// "sigs.k8s.io/prow/pkg/config/secret"
	// "sigs.k8s.io/prow/pkg/flagutil"
	"sigs.k8s.io/prow/pkg/git/v2"
	// "sigs.k8s.io/prow/pkg/github"
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
