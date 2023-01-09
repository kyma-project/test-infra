package client

import (
	"flag"
	"fmt"

	prowflagutil "k8s.io/test-infra/prow/flagutil"
	"k8s.io/test-infra/prow/github"
)

const (
	SapToolsHost            = "github.tools.sap"
	SapToolsGraphQLEndpoint = "https://github.tools.sap"
)

// GithubClientConfig holds configuration for GithubClient.
type GithubClientConfig struct {
	prowflagutil.GitHubOptions
	token  string
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
type GithubClientOption func(*flag.FlagSet) error

// NewGithubClient is a constructor function for GithubClient.
// A constructed client can be configured by providing GithubClientOptions.
func (o *GithubClientConfig) NewGithubClient(options ...GithubClientOption) (GithubClient, error) {
	var err error
	// Add flag to default github client flag set.
	// New flag enable client to authenticate with token string.
	fs := flag.NewFlagSet("ghclient", flag.ExitOnError)
	fs.StringVar(&o.token, "github-token", "", "Github Token")
	o.AddFlags(fs)
	// Run provided configuration option functions.
	for _, opt := range options {
		err = opt(fs)
		if err != nil {
			return nil, fmt.Errorf("failed applying functional option: %w", err)
		}
	}
	err = fs.Parse([]string{})
	if err != nil {
		return nil, err
	}
	var client github.Client
	if o.token != "" {
		client = o.GitHubClientWithAccessToken(o.token)
	} else {
		client, err = o.GitHubOptions.GitHubClient(o.DryRun)
		if err != nil {
			return nil, err
		}
	}
	return &githubClient{Client: client}, nil
}

// WithTokenPath is a client constructor configuration option passing path to a file with GitHub token.
func WithTokenPath(tokenpath string) GithubClientOption {
	return func(fs *flag.FlagSet) error {
		return fs.Set("github-token-path", tokenpath)
	}
}

// WithToken is a client constructor configuration option passing a GitHub token to authenticate.
func WithToken(token string) GithubClientOption {
	return func(fs *flag.FlagSet) error {
		return fs.Set("github-token", token)
	}
}

// WithEndpoint is a client constructor configuration option passing a GitHub endpoint.
func WithEndpoint(endpoint string) GithubClientOption {
	return func(fs *flag.FlagSet) error {
		return fs.Set("github-endpoint", endpoint)
	}
}

// WithHost is a client constructor configuration option passing a GitHub host.
func WithHost(hostname string) GithubClientOption {
	return func(fs *flag.FlagSet) error {
		return fs.Set("github-host", hostname)
	}
}
