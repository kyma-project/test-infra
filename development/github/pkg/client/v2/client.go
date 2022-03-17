package client

import (
	"fmt"

	prowflagutil "k8s.io/test-infra/prow/flagutil"
	"k8s.io/test-infra/prow/github"
)

type GithubClientConfig struct {
	prowflagutil.GitHubOptions
	DryRun bool
}

type GithubClient struct {
	github.Client
}

type GithubClientOption func(*GithubClientConfig) error

func NewGithubClient(options ...GithubClientOption) (*GithubClient, error) {
	conf := &GithubClientConfig{
		GitHubOptions: prowflagutil.GitHubOptions{},
		DryRun:        false,
	}

	for _, opt := range options {
		err := opt(conf)
		if err != nil {
			return nil, fmt.Errorf("failed applying functional option: %w", err)
		}
	}
	client, err := conf.GitHubOptions.GitHubClient(conf.DryRun)
	if err != nil {
		return nil, err
	}
	return &GithubClient{
		Client: client,
	}, nil
}

func WithTokenPath(tokenpath string) GithubClientOption {
	return func(conf *GithubClientConfig) error {
		conf.TokenPath = tokenpath
		return nil
	}
}
