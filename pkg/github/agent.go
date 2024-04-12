package github

import (
	"fmt"
	"github.com/kyma-project/test-infra/pkg/github/client/v2"
	"github.com/kyma-project/test-infra/pkg/github/repoowners"

	"github.com/kyma-project/test-infra/pkg/logging"
	"sigs.k8s.io/prow/prow/git/v2"
)

// ClientsAgent group clients used to interact with GitHub.
type ClientsAgent struct {
	logging.LoggerInterface
	GithubClient *client.GithubClient
	GitClient    *git.ClientFactory
	OwnersClient *repoowners.OwnersClient
}

// AgentOption is an agent constructor configuration option passing configuration to the agent constructor.
type AgentOption func(*ClientsAgent) error

// NewClientsAgent is a constructor function for ClientsAgent.
// A constructed agent can be configured by providing AgentOptions.
func NewClientsAgent(options ...AgentOption) (*ClientsAgent, error) {
	ca := &ClientsAgent{
		GithubClient: nil,
		GitClient:    nil,
		OwnersClient: nil,
	}

	for _, opt := range options {
		err := opt(ca)
		if err != nil {
			return nil, fmt.Errorf("failed applying functional option: %w", err)
		}
	}

	return ca, nil
}

// WithLogger is an agent constructor configuration option passing logger instance.
func WithLogger(logger *logging.LoggerInterface) AgentOption {
	return func(ca *ClientsAgent) error {
		ca.LoggerInterface = *logger
		return nil
	}
}

// WithGithubClient is an agent constructor configuration option passing a GitHub client instance.
func WithGithubClient(githubClient *client.GithubClient) AgentOption {
	return func(ca *ClientsAgent) error {
		ca.GithubClient = githubClient
		return nil
	}
}

// WithGitClient is an agent constructor configuration option passing a Git client instance.
func WithGitClient(gitClient *git.ClientFactory) AgentOption {
	return func(ca *ClientsAgent) error {
		ca.GitClient = gitClient
		return nil
	}
}

// WithOwnersClient is an agent constructor configuration option passing a repoowners client instance.
func WithOwnersClient(ownersClient *repoowners.OwnersClient) AgentOption {
	return func(ca *ClientsAgent) error {
		ca.OwnersClient = ownersClient
		return nil
	}
}
