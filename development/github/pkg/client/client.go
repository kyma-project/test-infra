package client

import (
	"context"
	"fmt"
	"github.com/google/go-github/v31/github"
	"github.com/kyma-project/test-infra/development/types"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
)

const (
	SapToolsGithubURL  = "https://github.tools.sap/"
	ProwGithubProxyURL = "http://ghproxy"
)

type SapToolsClient struct {
	*Client
}

type Client struct {
	*github.Client
}

func newOauthHttpClient(ctx context.Context, accessToken string) *http.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: accessToken,
			TokenType:   "token",
		},
	)

	return oauth2.NewClient(ctx, ts)
}

// TODO: create client with prow ghproxy as endpoint.
func NewClient(ctx context.Context, accessToken string) (*Client, error) {
	tc := newOauthHttpClient(ctx, accessToken)
	c := github.NewClient(tc)

	return &Client{Client: c}, nil
}

func NewSapToolsClient(ctx context.Context, accessToken string) (*SapToolsClient, error) {
	tc := newOauthHttpClient(ctx, accessToken)
	c, err := github.NewEnterpriseClient(SapToolsGithubURL, SapToolsGithubURL, tc)
	if err != nil {
		return nil, fmt.Errorf("got error when creating sap tools github enterprise client: %w", err)
	}

	return &SapToolsClient{Client: &Client{Client: c}}, nil
}

//func (c *Client) GetPrAuthorLogin(ctx context.Context, prNumber int, repoName, repoOwner string) (string, error) {
//	pr, resp, err := c.PullRequests.Get(ctx, repoOwner, repoName, prNumber)
//	return pr.GetUser().GetLogin(), nil
//}

func (c *SapToolsClient) GetUsersMap(ctx context.Context) ([]types.User, error) {
	var usersMap []types.User
	usersMapFile, _, resp, err := c.Client.Repositories.GetContents(ctx, "kyma", "test-infra", "/users-map.yaml", &github.RepositoryContentGetOptions{Ref: "main"})
	if err != nil {
		return nil, fmt.Errorf("got error when getting users-map.yaml file from github.tools.sap, error: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("got error when reading response body for non 200 HTTP reponse code, error: %w", err)
		}
		bodyString := string(bodyBytes)
		return nil, fmt.Errorf("got non 200 response code when getting users-map.taml file from github.tools.sap, body: %w", bodyString)
	}
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
