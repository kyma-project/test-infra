package client

import (
	"fmt"

	prowflagutil "sigs.k8s.io/prow/prow/flagutil"
	"sigs.k8s.io/prow/prow/github"
)

// GithubClientConfig holds configuration for GithubClient.
type GithubClientConfig struct {
	prowflagutil.GitHubOptions
	DryRun bool
}

// GithubClient is an implementation of GitHub client wrapping k8s test-infra GitHub Client.
type GithubClient interface {
	github.Client
}

// githubClient is an implementation of GitHub client wrapping k8s test-infra GitHub Client.
type githubClient struct {
	github.Client
}

// GithubClientOption is a client constructor configuration option passing configuration to the client constructor.
type GithubClientOption func(*GithubClientConfig) error

// NewGithubClient is a constructor function for GithubClient.
// A constructed client can be configured by providing GithubClientOptions.
func (o *GithubClientConfig) NewGithubClient(options ...GithubClientOption) (GithubClient, error) {
	// Run provided configuration option functions.
	for _, opt := range options {
		err := opt(o)
		if err != nil {
			return nil, fmt.Errorf("failed applying functional option: %w", err)
		}
	}
	client, err := o.GitHubOptions.GitHubClient(o.DryRun)
	if err != nil {
		return nil, err
	}
	return &githubClient{Client: client}, nil
}

// WithTokenPath is a client constructor configuration option passing path to a file with GitHub token.
func WithTokenPath(tokenpath string) GithubClientOption {
	return func(o *GithubClientConfig) error {
		o.TokenPath = tokenpath
		return nil
	}
}
