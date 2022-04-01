package repoowners

import (
	"flag"
	"fmt"

	"github.com/kyma-project/test-infra/development/github/pkg/client/v2"
	"github.com/kyma-project/test-infra/development/logging"
	"k8s.io/test-infra/prow/config"
	configflagutil "k8s.io/test-infra/prow/flagutil/config"
	pluginsflagutil "k8s.io/test-infra/prow/flagutil/plugins"
	"k8s.io/test-infra/prow/git/v2"
	"k8s.io/test-infra/prow/plugins"
	"k8s.io/test-infra/prow/repoowners"
)

type OwnersClientConfig struct {
	pluginsConfig pluginsflagutil.PluginOptions
	prowConfig    configflagutil.ConfigOptions
	gitClient     git.ClientFactory
	githubClient  *client.GithubClient
	logger        logging.LoggerInterface
}

func (o *OwnersClientConfig) AddFlags(fs *flag.FlagSet) {
	o.pluginsConfig.PluginConfigPathDefault = "/etc/plugins/plugins.yaml"
	o.prowConfig.AddFlags(fs)
	o.pluginsConfig.AddFlags(fs)
}

type OwnersClient struct {
	*repoowners.Client
	PluginsConfigAgent *plugins.ConfigAgent
	configAgent        *config.Agent
	Logger             logging.LoggerInterface
}

type RepoOwnersClientOption func(*OwnersClientConfig) error

func (o *OwnersClientConfig) NewRepoOwnersClient(options ...RepoOwnersClientOption) (*OwnersClient, error) {
	var err error
	repoOwnersClient := &OwnersClient{}

	for _, opt := range options {
		err := opt(o)
		if err != nil {
			return nil, fmt.Errorf("failed applying functional option: %w", err)
		}
	}

	if o.gitClient == nil {
		return nil, fmt.Errorf("git client not provided")
	}
	if o.githubClient == nil {
		return nil, fmt.Errorf("github client not provided")
	}
	if o.logger == nil {
		repoOwnersClient.Logger = o.logger
	} else {
		repoOwnersClient.Logger = logging.NewLogger()
	}

	repoOwnersClient.configAgent, err = o.prowConfig.ConfigAgent()
	if err != nil {
		return nil, fmt.Errorf("failed starting prow configuration agent, error: %w", err)
	}

	repoOwnersClient.PluginsConfigAgent, err = o.pluginsConfig.PluginAgent()
	if err != nil {
		return nil, fmt.Errorf("failed starting plugins configuration agent, error: %w", err)
	}

	ownersDirDenylist := func() *config.OwnersDirDenylist {
		// OwnersDirDenylist struct contains some defaults that's required by all
		// repos, so this function cannot return nil
		res := &config.OwnersDirDenylist{}
		deprecated := repoOwnersClient.configAgent.Config().OwnersDirBlacklist
		if l := repoOwnersClient.configAgent.Config().OwnersDirDenylist; l != nil {
			res = l
		}
		if deprecated != nil {
			repoOwnersClient.Logger.Warn("owners_dir_blacklist will be deprecated after October 2021, use owners_dir_denylist instead")
			if res != nil {
				repoOwnersClient.Logger.Warn("Both owners_dir_blacklist and owners_dir_denylist are provided, owners_dir_blacklist is discarded")
			} else {
				res = deprecated
			}
		}
		return res
	}

	repoOwnersClient.Client = repoowners.NewClient(o.gitClient,
		o.githubClient,
		repoOwnersClient.PluginsConfigAgent.Config().MDYAMLEnabled,
		repoOwnersClient.PluginsConfigAgent.Config().SkipCollaborators,
		ownersDirDenylist,
		repoOwnersClient.PluginsConfigAgent.Config().OwnersFilenames)
	return repoOwnersClient, nil
}

func WithGitClient(gitClient git.ClientFactory) RepoOwnersClientOption {
	return func(o *OwnersClientConfig) error {
		if o.gitClient == nil {
			o.gitClient = gitClient
			return nil
		} else {
			return fmt.Errorf("git client already defined")
		}
	}
}

func WithGithubClient(githubClient *client.GithubClient) RepoOwnersClientOption {
	return func(o *OwnersClientConfig) error {
		if o.githubClient == nil {
			o.githubClient = githubClient
			return nil
		} else {
			return fmt.Errorf("github client already defined")
		}
	}
}

func WithLogger(logger logging.LoggerInterface) RepoOwnersClientOption {
	return func(o *OwnersClientConfig) error {
		o.logger = logger
		return nil
	}
}
