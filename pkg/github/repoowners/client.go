package repoowners

import (
	"flag"
	"fmt"
	"github.com/kyma-project/test-infra/pkg/github/client/v2"

	"github.com/kyma-project/test-infra/pkg/logging"
	"sigs.k8s.io/prow/prow/config"
	configflagutil "sigs.k8s.io/prow/prow/flagutil/config"
	pluginsflagutil "sigs.k8s.io/prow/prow/flagutil/plugins"
	"sigs.k8s.io/prow/prow/git/v2"
	"sigs.k8s.io/prow/prow/plugins"
	"sigs.k8s.io/prow/prow/repoowners"
)

// OwnersClientConfig holds configuration for OwnersClient.
type OwnersClientConfig struct {
	pluginsConfig pluginsflagutil.PluginOptions
	prowConfig    configflagutil.ConfigOptions
	gitClient     git.ClientFactory
	githubClient  *client.GithubClient
	logger        logging.LoggerInterface
}

// AddFlags add OwnersClient flags to provided flag set.
// This allow to parse flags with flags provided by other components.
func (o *OwnersClientConfig) AddFlags(fs *flag.FlagSet) {
	o.pluginsConfig.PluginConfigPathDefault = "/etc/plugins/plugins.yaml"
	o.prowConfig.AddFlags(fs)
	o.pluginsConfig.AddFlags(fs)
}

// OwnersClient is a wrapper for k8s repoowners Client.
// It provides console loger as default.
type OwnersClient struct {
	*repoowners.Client
	PluginsConfigAgent *plugins.ConfigAgent
	configAgent        *config.Agent
	Logger             logging.LoggerInterface
}

// ClientOption is a client constructor configuration option passing configuration to the constructor.
type ClientOption func(*OwnersClientConfig) error

// NewRepoOwnersClient is a constructor of OwnersClient. Client can be configured with ClientOption.
// It provides console logger a default instance.
func (o *OwnersClientConfig) NewRepoOwnersClient(options ...ClientOption) (*OwnersClient, error) {
	var err error
	repoOwnersClient := &OwnersClient{}

	// Run provided ClientOption.
	for _, opt := range options {
		err := opt(o)
		if err != nil {
			return nil, fmt.Errorf("failed applying functional option: %w", err)
		}
	}

	// Check if mandatory field is provided.
	if o.gitClient == nil {
		return nil, fmt.Errorf("git client not provided")
	}
	if o.githubClient == nil {
		return nil, fmt.Errorf("github client not provided")
	}
	// Create default logger if no logger instance provided.
	if o.logger == nil {
		repoOwnersClient.Logger = o.logger
	} else {
		repoOwnersClient.Logger = logging.NewLogger()
	}

	// Use prow configuration.
	repoOwnersClient.configAgent, err = o.prowConfig.ConfigAgent()
	if err != nil {
		return nil, fmt.Errorf("failed starting prow configuration agent, error: %w", err)
	}

	// User prow plugin configuration.
	repoOwnersClient.PluginsConfigAgent, err = o.pluginsConfig.PluginAgent()
	if err != nil {
		return nil, fmt.Errorf("failed starting plugins configuration agent, error: %w", err)
	}

	ownersDirDenylist := func() *config.OwnersDirDenylist {
		// OwnersDirDenylist struct contains some defaults that's required by all
		// repos, so this function cannot return nil
		res := &config.OwnersDirDenylist{}
		deprecated := repoOwnersClient.configAgent.Config().OwnersDirDenylist
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

	// Create client
	repoOwnersClient.Client = repoowners.NewClient(o.gitClient,
		*o.githubClient,
		repoOwnersClient.PluginsConfigAgent.Config().MDYAMLEnabled,
		repoOwnersClient.PluginsConfigAgent.Config().SkipCollaborators,
		ownersDirDenylist,
		repoOwnersClient.PluginsConfigAgent.Config().OwnersFilenames)
	return repoOwnersClient, nil
}

// WithGitClient is constructor function configuration option providing git client.
func WithGitClient(gitClient git.ClientFactory) ClientOption {
	return func(o *OwnersClientConfig) error {
		if o.gitClient == nil {
			o.gitClient = gitClient
			return nil
		}
		return fmt.Errorf("git client already defined")

	}
}

// WithGithubClient is constructor function configuration option providing GitHub client.
func WithGithubClient(githubClient *client.GithubClient) ClientOption {
	return func(o *OwnersClientConfig) error {
		if o.githubClient == nil {
			o.githubClient = githubClient
			return nil
		}
		return fmt.Errorf("github client already defined")
	}
}

// WithLogger is constructor function configuration option providing logger instance.
func WithLogger(logger logging.LoggerInterface) ClientOption {
	return func(o *OwnersClientConfig) error {
		o.logger = logger
		return nil
	}
}
