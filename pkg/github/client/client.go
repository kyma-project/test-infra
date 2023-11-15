package client

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"flag"
	"fmt"
	"github.com/kyma-project/test-infra/pkg/gcp/cloudfunctions"
	"github.com/kyma-project/test-infra/pkg/types"
	"hash"
	"net/http"
	"sync"

	"github.com/google/go-github/v48/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
	"k8s.io/test-infra/prow/config/secret"
)

const (
	SapToolsGithubURL  = "https://github.tools.sap/"
	ProwGithubProxyURL = "http://ghproxy"
)

// GithubClientConfig holds configuration for GitHub client.
type GithubClientConfig struct {
	tokenPath tokenPathFlag
}

// tokenPathFlag is used as cli flag. When set it adds path to secret agent.
type tokenPathFlag string

// SapToolsClient wraps kyma implementation github Client and provides additional methods.
type SapToolsClient struct {
	*Client
}

// Client wraps google github Client and provides additional methods.
type Client struct {
	*github.Client
	hmacKey []byte // A random generated key for hmac hashing.
	// Token hmac hash used to authenticate client on GitHub.
	// Before reauthenticating client, check if new token is different from stored token hmac hash.
	tokenHmac hash.Hash
	// Used to prevent race condition when reauthenticating client.
	// RLock and RUnlock must be used to secure all client methods calls.
	WrapperClientMu sync.RWMutex
}

// String provide string representation for tokenPathFlag.
// This is to implement flag.Var interface.
func (f *tokenPathFlag) String() string {
	return string(*f)
}

// Set implement setting value from cli for tokenPathFlag.
// This is to implement flag.Var interface.
// Set will add path to token to a secret agent.
// Set provide default value for flag.
func (f *tokenPathFlag) Set(value string) error {
	if value == "" {
		value = "/etc/v1github/oauth"
	}
	*f = tokenPathFlag(value)
	// Add path to secret agent.
	return secret.Add(value)
}

// AddFlags add GitHub Client cli flag to provided flag set.
// It lets parse flags with flags provided by other components.
func (o *GithubClientConfig) AddFlags(fs *flag.FlagSet) {
	fs.Var(&o.tokenPath, "v1-github-token-path", "Environment variable name with github token.")
}

// GetToken retrieve GitHub token from secret agent.
// It use a token path set on Github Client.
func (o *GithubClientConfig) GetToken() (string, error) {
	token := secret.GetSecret(string(o.tokenPath))
	if string(token) == "" {
		return "", fmt.Errorf("tools GitHub token is empty")
	}
	return string(token), nil
}

// newOauthHTTPClient creates HTTP client with oauth authentication.
// It authenticates with Bearer token.
func newOauthHTTPClient(ctx context.Context, accessToken string) *http.Client {
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
	tc := newOauthHTTPClient(ctx, accessToken)
	ghc := github.NewClient(tc)
	c := &Client{
		Client:          ghc,
		WrapperClientMu: sync.RWMutex{},
	}
	err := c.generateHmacKey()
	if err != nil {
		return nil, err
	}
	err = c.storeTokenHash([]byte(accessToken))
	if err != nil {
		return nil, err
	}
	return c, nil
}

// NewSapToolsClient creates kyma implementation github Client with SapToolsGithubURL as an endpoint.
// Client uses oauth authentication with bearer token.
func NewSapToolsClient(ctx context.Context, accessToken string) (*SapToolsClient, error) {
	tc := newOauthHTTPClient(ctx, accessToken)
	ghec, err := github.NewEnterpriseClient(SapToolsGithubURL, SapToolsGithubURL, tc)
	if err != nil {
		return nil, err
	}
	c := &Client{
		Client:          ghec,
		WrapperClientMu: sync.RWMutex{},
	}
	err = c.generateHmacKey()
	if err != nil {
		return nil, err
	}
	err = c.storeTokenHash([]byte(accessToken))
	if err != nil {
		return nil, err
	}

	return &SapToolsClient{c}, nil
}

// IsStatusOK will check if http response code is 200.
// On not OK status it will read response body to expose details about error.
func (c *Client) IsStatusOK(resp *github.Response) (bool, error) {
	return IsStatusOK(resp)
}

// IsStatusOK will check if http response code is in 2xx range.
func IsStatusOK(resp *github.Response) (bool, error) {
	statusOK := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !statusOK {
		return false, fmt.Errorf("got %d response code in HTTP response", resp.StatusCode)
	}
	return true, nil
}

// generateHmacKey generate cryptographically safe random key.
// Key is stored in a Client struct and used to hash password with hmac.
func (c *Client) generateHmacKey() error {
	buf := make([]byte, 128)
	_, err := rand.Read(buf)
	if err != nil {
		return err
	}
	c.hmacKey = buf
	return nil
}

// storeTokenHash hash token with hmac sha256 and store it in a client.
// Using hmac as it's designed for passwords secure storage.
func (c *Client) storeTokenHash(token []byte) error {
	mac := hmac.New(sha256.New, c.hmacKey)
	_, err := mac.Write(token)
	if err != nil {
		return err
	}
	c.tokenHmac = mac
	return nil
}

// CompareTokensHashes generates hmac hash for provided token and compare it with hash stored in a client.
// If provided token is the same as a token stored in a client return nil hash.Hash and error, if not new token hmac hash
// and nil error is returned.
// Compare tokens to check if client reauthentication with new token is required.
// Creating a new client with wrong token doesn't return error.
// A client will get error again on first GitHub API call, so check first if new password exist and reauthentication
// is required.
func (c *Client) compareTokensHashes(token []byte) (hash.Hash, error) {
	mac := hmac.New(sha256.New, c.hmacKey)
	_, err := mac.Write(token)
	if err != nil {
		return nil, err
	}
	if eq := hmac.Equal(c.tokenHmac.Sum(nil), mac.Sum(nil)); eq {
		return nil, nil
	}
	return mac, nil
}

// Reauthenticate creates new GitHub Enterprise client with provided access token and replace existing ones.
// It locks wrapper client mutex for read and write to prevent race condition between client threads.
// A caller should retry failed GitHub API call on non error Reauthenticate execution.
// Because multiple threads may wait to reauthenticate and second or later thread will not detect a token change,
// method will not raise error and log a warning message. This is to let caller to retry a GitHub API call.
// TODO: replace cloudfunctions logger with interface
func (c *SapToolsClient) Reauthenticate(ctx context.Context, logger *cloudfunctions.LogEntry, accessToken []byte) (bool, error) {
	c.WrapperClientMu.Lock()
	defer c.WrapperClientMu.Unlock()
	tokenHmac, err := c.compareTokensHashes(accessToken)
	if err != nil {
		logger.LogCritical("failed compare token hashes, error %s", err)
	}
	if tokenHmac == nil {
		logger.LogWarning("No new token available for GitHub client, can't reauthenticate.")
		return false, nil
	}
	tc := newOauthHTTPClient(ctx, string(accessToken))
	ghec, err := github.NewEnterpriseClient(SapToolsGithubURL, SapToolsGithubURL, tc)
	if err != nil {
		return false, err
	}
	c.Client.Client = ghec
	c.tokenHmac = tokenHmac
	logger.LogInfo("New token provided, updated client with new credentials.")
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
