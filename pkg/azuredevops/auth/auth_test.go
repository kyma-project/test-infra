package auth

import (
	"testing"
)

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
