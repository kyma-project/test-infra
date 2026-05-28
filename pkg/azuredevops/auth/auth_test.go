package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

type mockCredential struct {
	token string
	err   error
}

func (m *mockCredential) GetToken(_ context.Context, _ policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: m.token, ExpiresOn: time.Now().Add(time.Hour)}, m.err
}

func TestServicePrincipalProvider_GetToken(t *testing.T) {
	tests := []struct {
		name      string
		mock      *mockCredential
		wantToken string
		wantErr   bool
	}{
		{
			name:      "returns token on success",
			mock:      &mockCredential{token: "my-bearer-token"},
			wantToken: "my-bearer-token",
			wantErr:   false,
		},
		{
			name:    "returns error on credential failure",
			mock:    &mockCredential{err: errors.New("auth failed")},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &ServicePrincipalProvider{cred: tt.mock}
			got, err := provider.GetToken(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("GetToken() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.wantToken {
				t.Errorf("GetToken() = %v, want %v", got, tt.wantToken)
			}
		})
	}
}

func TestGetToken(t *testing.T) {
	tests := []struct {
		name      string
		mock      *mockCredential
		wantToken string
		wantErr   bool
	}{
		{
			name:      "returns token on success",
			mock:      &mockCredential{token: "my-bearer-token"},
			wantToken: "my-bearer-token",
			wantErr:   false,
		},
		{
			name:    "returns error on credential failure",
			mock:    &mockCredential{err: errors.New("auth failed")},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getToken(context.Background(), tt.mock)
			if (err != nil) != tt.wantErr {
				t.Errorf("getToken() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.wantToken {
				t.Errorf("getToken() = %v, want %v", got, tt.wantToken)
			}
		})
	}
}

func TestServicePrincipalConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     ServicePrincipalConfig
		wantErr bool
	}{
		{
			name:    "valid config",
			cfg:     ServicePrincipalConfig{TenantID: "tenant", ClientID: "client", ClientSecret: "secret"},
			wantErr: false,
		},
		{
			name:    "missing TenantID",
			cfg:     ServicePrincipalConfig{ClientID: "client", ClientSecret: "secret"},
			wantErr: true,
		},
		{
			name:    "missing ClientID",
			cfg:     ServicePrincipalConfig{TenantID: "tenant", ClientSecret: "secret"},
			wantErr: true,
		},
		{
			name:    "missing ClientSecret",
			cfg:     ServicePrincipalConfig{TenantID: "tenant", ClientID: "client"},
			wantErr: true,
		},
		{
			name:    "empty config",
			cfg:     ServicePrincipalConfig{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewServicePrincipalProvider(t *testing.T) {
	tests := []struct {
		name    string
		cfg     ServicePrincipalConfig
		wantErr bool
	}{
		{
			name:    "valid config returns provider",
			cfg:     ServicePrincipalConfig{TenantID: "tenant", ClientID: "client", ClientSecret: "secret"},
			wantErr: false,
		},
		{
			name:    "invalid config returns error",
			cfg:     ServicePrincipalConfig{},
			wantErr: true,
		},
		{
			name:    "partial config returns error",
			cfg:     ServicePrincipalConfig{TenantID: "tenant"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewServicePrincipalProvider(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewServicePrincipalProvider() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && provider == nil {
				t.Error("NewServicePrincipalProvider() returned nil provider without error")
			}
		})
	}
}
