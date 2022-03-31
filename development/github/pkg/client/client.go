package client

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/go-github/v42/github"
	"github.com/kyma-project/test-infra/development/types"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

const (
	SapToolsGithubURL  = "https://github.tools.sap/"
	ProwGithubProxyURL = "http://ghproxy"
)

// SapToolsClient wraps kyma implementation github Client and provides additional methods.
type SapToolsClient struct {
	*Client
}

// Client wraps google github Client and provides additional methods.
type Client struct {
	*github.Client
}

// newOauthHttpClient creates HTTP client with oauth authentication.
// It authenticates with Bearer token.
func newOauthHttpClient(ctx context.Context, accessToken string) *http.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: accessToken,
			TokenType:   "token",
		},
	)

	return oauth2.NewClient(ctx, ts)
}

// NewClient creates kyma implementation of github client with oauth authentication.
// TODO: create client with support for github cache or ghproxy.
func NewClient(ctx context.Context, accessToken string) (*Client, error) {
	tc := newOauthHttpClient(ctx, accessToken)
	c := github.NewClient(tc)

	return &Client{c}, nil
}

// NewSapToolsClient creates kyma implementation github Client with SapToolsGithubURL as an endpoint.
// Client uses oauth authentication with bearer token.
func NewSapToolsClient(ctx context.Context, accessToken string) (*SapToolsClient, error) {
	tc := newOauthHttpClient(ctx, accessToken)
	c, err := github.NewEnterpriseClient(SapToolsGithubURL, SapToolsGithubURL, tc)
	if err != nil {
		return nil, fmt.Errorf("got error when creating sap tools github enterprise client: %w", err)
	}

	return &SapToolsClient{&Client{c}}, nil
}

// IsStatusOK will check if http response code is 200.
// On not OK status it will read response body to expose details about error.
func (c *Client) IsStatusOK(resp *github.Response) (bool, error) {
	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return false, fmt.Errorf("got error when reading response body for non 200 HTTP reponse code, error: %w", err)
		}
		bodyString := string(bodyBytes)
		return false, fmt.Errorf("got non 200 response code in HTTP response, body: %s", bodyString)
	}
	return true, nil
}

// GetUsersMap will get users-map.yaml file from github.tools.sap/kyma/test-infra repository.
func (c *SapToolsClient) GetUsersMap(ctx context.Context) ([]types.User, error) {
	var usersMap []types.User
	// Get file from github.
	usersMapFile, _, resp, err := c.Repositories.GetContents(ctx, "kyma", "test-infra", "/users-map.yaml", &github.RepositoryContentGetOptions{Ref: "main"})
	if err != nil {
		return nil, fmt.Errorf("got error when getting users-map.yaml file from github.tools.sap, error: %w", err)
	}
	// Check HTTP response code
	if ok, err := c.IsStatusOK(resp); !ok {
		return nil, err
	}
	// Read file content.
	usersMapString, err := usersMapFile.GetContent()
	if err != nil {
		return nil, fmt.Errorf("got error when getting content of users-map.yaml file, error: %w", err)
	}
	err = yaml.Unmarshal([]byte(usersMapString), &usersMap)
	if err != nil {
		return nil, fmt.Errorf("got error when unmarshaling usres-map.yaml file content, error: %w", err)
	}
	return usersMap, nil
}

// GetAliasesMap will get aliasess-map.yaml file from github.tools.sap/kyma/test-infra repository.
func (c *SapToolsClient) GetAliasesMap(ctx context.Context) ([]types.Alias, error) {
	var aliasesMap []types.Alias
	// Get file from github.
	aliasesMapFile, _, resp, err := c.Repositories.GetContents(ctx, "kyma", "test-infra", "/aliases-map.yaml", &github.RepositoryContentGetOptions{Ref: "main"})
	if err != nil {
		return nil, fmt.Errorf("got error when getting users-map.yaml file from github.tools.sap, error: %w", err)
	}
	// Check HTTP response code
	if ok, err := c.IsStatusOK(resp); !ok {
		return nil, err
	}
	// Read file content.
	aliasesMapString, err := aliasesMapFile.GetContent()
	if err != nil {
		return nil, fmt.Errorf("got error when getting content of users-map.yaml file, error: %w", err)
	}
	err = yaml.Unmarshal([]byte(aliasesMapString), &aliasesMap)
	if err != nil {
		return nil, fmt.Errorf("got error when unmarshaling usres-map.yaml file content, error: %w", err)
	}
	return aliasesMap, nil
}

// GetAuthorLoginForBranch will provide commit author github Login for given SHA.
func (c *Client) GetAuthorLoginForBranch(ctx context.Context, branchName, owner, repo string) (*string, error) {
	branch, resp, err := c.Repositories.GetBranch(ctx, owner, repo, branchName, true)
	if err != nil {
		return nil, fmt.Errorf("got error when getting commit, error: %w", err)
	}
	// Check HTTP response code.
	if ok, err := c.IsStatusOK(resp); !ok {
		return nil, err
	}
	commit := branch.GetCommit()
	// Read commit author Login.
	l := commit.GetAuthor().GetLogin()
	return &l, nil
}

// GetAuthorLoginForSHA will provide commit author github Login for given SHA.
func (c *Client) GetAuthorLoginForSHA(ctx context.Context, sha, owner, repo string) (*string, error) {
	// Get commit for SHA.
	commit, resp, err := c.Repositories.GetCommit(ctx, owner, repo, sha, nil)
	if err != nil {
		return nil, fmt.Errorf("got error when getting commit, error: %w", err)
	}
	// Check HTTP response code.
	if ok, err := c.IsStatusOK(resp); !ok {
		return nil, err
	}
	// Read commit author Login.
	l := commit.GetAuthor().GetLogin()
	return &l, nil
}
