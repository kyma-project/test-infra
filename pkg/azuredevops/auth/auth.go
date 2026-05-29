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
	cred azcore.TokenCredential
}

// NewServicePrincipalCredential creates an Azure AD ClientSecretCredential from the provided config.
func NewServicePrincipalCredential(cfg ServicePrincipalConfig) (azcore.TokenCredential, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid service principal config: %w", err)
	}
	cred, err := azidentity.NewClientSecretCredential(cfg.TenantID, cfg.ClientID, cfg.ClientSecret, nil)
	if err != nil {
		return nil, fmt.Errorf("failed creating service principal credential: %w", err)
	}
	return cred, nil
}

// NewServicePrincipalProvider creates a ServicePrincipalProvider with the provided credential.
func NewServicePrincipalProvider(cred azcore.TokenCredential) *ServicePrincipalProvider {
	return &ServicePrincipalProvider{cred: cred}
}

// GetToken acquires a Bearer token for Azure DevOps.
func (p *ServicePrincipalProvider) GetToken(ctx context.Context) (string, error) {
	token, err := p.cred.GetToken(ctx, policy.TokenRequestOptions{Scopes: []string{adoScope}})
	if err != nil {
		return "", fmt.Errorf("failed acquiring service principal token: %w", err)
	}
	return token.Token, nil
}
