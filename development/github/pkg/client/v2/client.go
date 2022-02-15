package client

import (
	"fmt"

	"github.com/kyma-project/test-infra/development/gcp/pkg/logging"
	"k8s.io/test-infra/prow/config"
	prowflagutil "k8s.io/test-infra/prow/flagutil"
	"k8s.io/test-infra/prow/git/v2"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/plugins"
	"k8s.io/test-infra/prow/plugins/ownersconfig"
	"k8s.io/test-infra/prow/repoowners"
)

type ClientsAgent struct {
	GithubConfig prowflagutil.GitHubOptions
	GitConfig    git.ClientFactoryOpts
	OwnersConfig OwnersConfig
	DryRun       bool
	Logger       *logging.Logger
	GithubClient github.Client
	GitClient    git.ClientFactory
	OwnersClient repoowners.Client
}

type OwnersConfig struct {
	plugins.Owners
	config.OwnersDirDenylist
}

func NewClientAgent() *ClientsAgent {
	return &ClientsAgent{}
}

func (c *ClientsAgent) WithLogger(logger *logging.Logger) *ClientsAgent {
	c.Logger = logger
	return c
}

func (c *ClientsAgent) WithGithubTokenPath(tokenpath string) *ClientsAgent {
	c.GithubConfig.TokenPath = tokenpath
	return c
}

func (c *ClientsAgent) WithGithubClient() (*ClientsAgent, error) {
	if c.GithubConfig.TokenPath == "" {
		return nil, fmt.Errorf("tokenPath value can't be empty string")
	}
	client, err := c.GithubConfig.GitHubClient(c.DryRun)
	if err != nil {
		return nil, err
	}
	c.GithubClient = client
	return c, nil
}

func (c *ClientsAgent) WithGitClient() (*ClientsAgent, error) {
	if c.GithubClient == nil {
		_, err := c.WithGithubClient()
		return nil, err
	}
	gitUser := func() (name, email string, err error) {
		user, err := c.GithubClient.BotUser()
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
	gitClient, err := git.NewClientFactory(opts.Apply)
	if err != nil {
		return nil, err
	}
	c.GitClient = gitClient
	return c, nil
}

func (c *ClientsAgent) WithOwnersClient() (*ClientsAgent, error) {
	if c.GitClient == nil {
		_, err := c.WithGitClient()
		if err != nil {
			return nil, err
		}
	}
	ownersClient := repoowners.NewClient(c.GitClient, c.GithubClient, c.mdYAMLEnabled, c.skipCollaborators, c.ownersDirDenylist, c.resolver)
	c.OwnersClient = *ownersClient
	return c, nil
}

func (c *ClientsAgent) mdYAMLEnabled(org, repo string) bool {
	full := fmt.Sprintf("%s/%s", org, repo)
	for _, elem := range c.OwnersConfig.MDYAMLRepos {
		if elem == org || elem == full {
			return true
		}
	}
	return false
}

func (c *ClientsAgent) skipCollaborators(org, repo string) bool {
	full := fmt.Sprintf("%s/%s", org, repo)
	for _, elem := range c.OwnersConfig.SkipCollaborators {
		if elem == org || elem == full {
			return true
		}
	}
	return false
}

func (c *ClientsAgent) ownersDirDenylist() *config.OwnersDirDenylist {
	// OwnersDirDenylist struct contains some defaults that's required by all
	// repos, so this function cannot return nil
	return &c.OwnersConfig.OwnersDirDenylist
}

func (c *ClientsAgent) resolver(org, repo string) ownersconfig.Filenames {
	// OwnersFilenames determines which filenames to use for OWNERS and OWNERS_ALIASES for a repo.
	full := fmt.Sprintf("%s/%s", org, repo)
	if filenames, configured := c.OwnersConfig.Filenames[full]; configured {
		return filenames
	}
	return ownersconfig.Filenames{
		Owners:        ownersconfig.DefaultOwnersFile,
		OwnersAliases: ownersconfig.DefaultOwnersAliasesFile,
	}
}
