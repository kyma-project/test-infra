package github

import (
	"fmt"

	"github.com/kyma-project/test-infra/development/gcp/pkg/logging"
	"github.com/kyma-project/test-infra/development/github/pkg/client/v2"
	"github.com/kyma-project/test-infra/development/github/pkg/repoowners"
	"k8s.io/test-infra/prow/git/v2"
)

type ClientsAgent struct {
	logging.LoggerInterface
	GithubClient *client.GithubClient
	GitClient    *git.ClientFactory
	OwnersClient *repoowners.RepoOwnersClient
}

type AgentOption func(*ClientsAgent) error

// TODO: With github, git, woners client
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

func WithLogger(logger *logging.LoggerInterface) AgentOption {
	return func(ca *ClientsAgent) error {
		ca.LoggerInterface = *logger
		return nil
	}
}
