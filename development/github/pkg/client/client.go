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
	SapToolsGithubURL = "https://github.tools.sap/"
)

type SapToolsClient struct {
	*github.Client
}

func NewSapToolsClient(ctx context.Context, accessToken string) (*SapToolsClient, error) {
	var sapToolsClient SapToolsClient
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: accessToken,
			TokenType:   "token",
		},
	)
	tc := oauth2.NewClient(ctx, ts)
	client, err := github.NewEnterpriseClient(SapToolsGithubURL, SapToolsGithubURL, tc)
	if err != nil {
		return nil, fmt.Errorf("got error when creating sap tools github enterprise client: %v", err)
	}
	sapToolsClient.Client = client
	return &sapToolsClient, nil
}

func (c *SapToolsClient) GetUsersMap() ([]types.User, error) {
	var usersMap []types.User
	usersMapFile, _, resp, err := c.Client.Repositories.GetContents(ctx, "kyma", "test-infra", "/users-map.yaml", &github.RepositoryContentGetOptions{Ref: "main"})
	if err != nil {
		return nil, fmt.Errorf("got error when getting users-map.yaml file from github.tools.sap, error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("got error when reading response body for non 200 HTTP reponse code, error: %v", err)
		}
		bodyString := string(bodyBytes)
		return nil, fmt.Errorf("got non 200 response code when getting users-map.taml file from github.tools.sap, body: %s", bodyString)
	}
	usersMapString, err := usersMapFile.GetContent()
	if err != nil {
		return nil, fmt.Errorf("got error when getting content of users-map.yaml file, error: %v", err)
	}
	err = yaml.Unmarshal([]byte(usersMapString), &usersMap)
	if err != nil {
		return nil, fmt.Errorf("got error when unmarshaling usres-map.yaml file content, error: %v", err)
	}
	return usersMap, nil
}
