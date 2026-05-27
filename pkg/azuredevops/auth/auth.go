package auth

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

// adoScope is the Azure DevOps resource ID used when requesting a token.
const adoScope = "499b84ac-1321-427f-aa17-267ca6975798/.default"

// TokenProvider acquires a Bearer token for use with Azure DevOps.
type TokenProvider interface {
	GetToken(ctx context.Context) (string, error)
}

// ServicePrincipalConfig holds Azure AD App Registration credentials.
type ServicePrincipalConfig struct {
	TenantID     string
	ClientID     string
	ClientSecret string
}

// Validate returns an error if any required field is empty.
func (c ServicePrincipalConfig) Validate() error {
	if c.TenantID == "" {
		return fmt.Errorf("TenantID is required")
	}
	if c.ClientID == "" {
		return fmt.Errorf("ClientID is required")
	}
	if c.ClientSecret == "" {
		return fmt.Errorf("ClientSecret is required")
	}
	return nil
}

// ServicePrincipalProvider implements TokenProvider using Azure AD Service Principal credentials.
type ServicePrincipalProvider struct {
	cfg ServicePrincipalConfig
}

// NewServicePrincipalProvider creates a ServicePrincipalProvider after validating the config.
func NewServicePrincipalProvider(cfg ServicePrincipalConfig) (*ServicePrincipalProvider, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid service principal config: %w", err)
	}
	return &ServicePrincipalProvider{cfg: cfg}, nil
}

// GetToken acquires a Bearer token for Azure DevOps.
func (p *ServicePrincipalProvider) GetToken(ctx context.Context) (string, error) {
	return GetServicePrincipalToken(ctx, p.cfg)
}

// GetServicePrincipalToken acquires a Bearer token for Azure DevOps using the provided Service Principal credentials.
func GetServicePrincipalToken(ctx context.Context, cfg ServicePrincipalConfig) (string, error) {
	if err := cfg.Validate(); err != nil {
		return "", fmt.Errorf("invalid service principal config: %w", err)
	}
	cred, err := azidentity.NewClientSecretCredential(cfg.TenantID, cfg.ClientID, cfg.ClientSecret, nil)
	if err != nil {
		return "", fmt.Errorf("failed creating service principal credential: %w", err)
	}
	return getToken(ctx, cred)
}

// getToken requests a token from the provided credential. Separated for testability.
func getToken(ctx context.Context, cred azcore.TokenCredential) (string, error) {
	token, err := cred.GetToken(ctx, policy.TokenRequestOptions{Scopes: []string{adoScope}})
	if err != nil {
		return "", fmt.Errorf("failed acquiring service principal token: %w", err)
	}
	return token.Token, nil
}
