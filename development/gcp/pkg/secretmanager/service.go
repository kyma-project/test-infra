package secretmanager

import (
	"context"
	"encoding/base64"
	"fmt"

	"google.golang.org/api/option"
	gcpsecretmanager "google.golang.org/api/secretmanager/v1"
)

type Service struct {
	Service        *gcpsecretmanager.Service
	VersionService *gcpsecretmanager.ProjectsSecretsVersionsService
}

// NewService creates new Service struct
func NewService(ctx context.Context, options ...option.ClientOption) (*Service, error) {
	secretManagerClient, err := gcpsecretmanager.NewService(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create google Secret Manager client, got error: %w", err)
	}
	secretVersionsService := gcpsecretmanager.NewProjectsSecretsVersionsService(secretManagerClient)

	return &Service{Service: secretManagerClient, VersionService: secretVersionsService}, nil
}

// AddSecretVersion adds a new version to a secret
// expects secretPath in "projects/*/secrets/*" format
func (sm *Service) AddSecretVersion(secretPath string, secretData []byte) (*gcpsecretmanager.SecretVersion, error) {
	newVersionRequest := gcpsecretmanager.AddSecretVersionRequest{Payload: &gcpsecretmanager.SecretPayload{Data: base64.StdEncoding.EncodeToString(secretData)}}
	newVersionCall := sm.Service.Projects.Secrets.AddVersion(secretPath, &newVersionRequest)
	secretVersion, err := newVersionCall.Do()
	return secretVersion, err
}

// ListSecretVersions lists all versions of a secret
// expects secretPath in "projects/*/secrets/*" format
func (sm *Service) ListSecretVersions(secretPath string) (*gcpsecretmanager.ListSecretVersionsResponse, error) {
	secretVersionsCall := sm.Service.Projects.Secrets.Versions.List(secretPath)
	secretVersions, err := secretVersionsCall.Do()
	return secretVersions, err
}

// GetSecretVersion retrieves one version of a secret
// expects secretPath in "projects/*/secrets/*/versions/*" format
func (sm *Service) GetSecretVersion(secretPath string) (*gcpsecretmanager.SecretVersion, error) {
	secretVersionsCall := sm.Service.Projects.Secrets.Versions.Get(secretPath)
	secretVersion, err := secretVersionsCall.Do()
	return secretVersion, err
}

// DisableSecretVersion disables a version of a secret
func (sm *Service) DisableSecretVersion(version *gcpsecretmanager.SecretVersion) (*gcpsecretmanager.SecretVersion, error) {
	disableRequest := gcpsecretmanager.DisableSecretVersionRequest{}
	disableCall := sm.VersionService.Disable(version.Name, &disableRequest)
	returnedSecretVersion, err := disableCall.Do()
	return returnedSecretVersion, err
}

// GetSecretVersionData retrieves payload of a secret version
// expects secretPath in "projects/*/secrets/*/versions/*" format
func (sm *Service) GetSecretVersionData(secretPath string) (string, error) {
	secretVersionCall := sm.VersionService.Access(secretPath)
	secretVersion, err := secretVersionCall.Do()
	if err != nil {
		return "", err
	}
	return secretVersion.Payload.Data, err
}
