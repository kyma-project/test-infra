package repoowners

import (
	"fmt"

	"github.com/kyma-project/test-infra/development/github/pkg/client/v2"
	"github.com/kyma-project/test-infra/development/github/pkg/git"
	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/plugins"
	"k8s.io/test-infra/prow/plugins/ownersconfig"
	"k8s.io/test-infra/prow/repoowners"
)

type OwnersClientConfig struct {
	*plugins.Owners
	config.OwnersDirDenylist
	gitClient    *git.GitClient
	githubClient *client.GithubClient
}

type OwnersClient struct {
	*repoowners.Client
}

type RepoOwnersClientOption func(*OwnersClientConfig) error

func NewRepoOwnersClient(options ...RepoOwnersClientOption) (*OwnersClient, error) {

	conf := &OwnersClientConfig{
		Owners:            nil,
		OwnersDirDenylist: config.OwnersDirDenylist{},
		gitClient:         nil,
		githubClient:      nil,
	}

	for _, opt := range options {
		err := opt(conf)
		if err != nil {
			return nil, fmt.Errorf("failed applying functional option: %w", err)
		}
	}

	if conf.gitClient == nil {
		return nil, fmt.Errorf("git client not provided")
	}
	if conf.githubClient == nil {
		return nil, fmt.Errorf("github client not provided")
	}
	repoOwnersClient := &OwnersClient{}

	repoOwnersClient.Client = repoowners.NewClient(conf.gitClient, conf.githubClient, conf.mdYAMLEnabled, conf.skipCollaborators, conf.ownersDirDenylist, conf.resolver)
	return repoOwnersClient, nil
}

func WithGitClient(gitClient *git.GitClient) RepoOwnersClientOption {
	return func(conf *OwnersClientConfig) error {
		if conf.gitClient == nil {
			conf.gitClient = gitClient
			return nil
		} else {
			return fmt.Errorf("git client already defined")
		}
	}
}

func WithGithubClient(githubClient *client.GithubClient) RepoOwnersClientOption {
	return func(conf *OwnersClientConfig) error {
		if conf.githubClient == nil {
			conf.githubClient = githubClient
			return nil
		} else {
			return fmt.Errorf("github client already defined")
		}
	}
}

func (c *OwnersClientConfig) mdYAMLEnabled(org, repo string) bool {
	full := fmt.Sprintf("%s/%s", org, repo)
	for _, elem := range c.MDYAMLRepos {
		if elem == org || elem == full {
			return true
		}
	}
	return false
}

func (c *OwnersClientConfig) skipCollaborators(org, repo string) bool {
	full := fmt.Sprintf("%s/%s", org, repo)
	for _, elem := range c.SkipCollaborators {
		if elem == org || elem == full {
			return true
		}
	}
	return false
}

func (c *OwnersClientConfig) ownersDirDenylist() *config.OwnersDirDenylist {
	// OwnersDirDenylist struct contains some defaults that's required by all
	// repos, so this function cannot return nil
	return &c.OwnersDirDenylist
}

func (c *OwnersClientConfig) resolver(org, repo string) ownersconfig.Filenames {
	// OwnersFilenames determines which filenames to use for OWNERS and OWNERS_ALIASES for a repo.
	full := fmt.Sprintf("%s/%s", org, repo)
	if filenames, configured := c.Filenames[full]; configured {
		return filenames
	}
	return ownersconfig.Filenames{
		Owners:        ownersconfig.DefaultOwnersFile,
		OwnersAliases: ownersconfig.DefaultOwnersAliasesFile,
	}
}
