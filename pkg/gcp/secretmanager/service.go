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
	return sm.Service.Projects.Secrets.AddVersion(secretPath, &newVersionRequest).Do()
}

// ListSecretVersions lists all versions of a secret
// expects secretPath in "projects/*/secrets/*" format
func (sm *Service) ListSecretVersions(secretPath string) (*gcpsecretmanager.ListSecretVersionsResponse, error) {
	return sm.Service.Projects.Secrets.Versions.List(secretPath).Do()
}

// GetLatestSecretVersion retrieves latest version of a secret
// expects secretPath in "projects/*/secrets/*" format
func (sm *Service) GetLatestSecretVersion(secretPath string) (*gcpsecretmanager.SecretVersion, error) {
	return sm.Service.Projects.Secrets.Versions.Get(secretPath + "/versions/latest").Do()
}

// GetSecretVersion retrieves one version of a secret
// expects secretPath in "projects/*/secrets/*/versions/*" format
func (sm *Service) GetSecretVersion(secretPath string) (*gcpsecretmanager.SecretVersion, error) {
	return sm.Service.Projects.Secrets.Versions.Get(secretPath).Do()
}

// GetSecretVersion retrieves one version of a secret
// expects secretPath in "projects/*/secrets/*/versions/*" format
func (sm *Service) GetAllSecretVersions(secretPath, filter string) ([]*gcpsecretmanager.SecretVersion, error) {
	var versions []*gcpsecretmanager.SecretVersion
	nextPageToken := ""

	for {
		allVersionsCall := sm.Service.Projects.Secrets.Versions.List(secretPath)

		if filter != "" {
			allVersionsCall = allVersionsCall.Filter(filter)
		}
		if nextPageToken != "" {
			allVersionsCall.PageToken(nextPageToken)
		}

		allVersionsResponse, err := allVersionsCall.Do()
		if err != nil {
			return nil, err
		}
		versions = append(versions, allVersionsResponse.Versions...)

		if allVersionsResponse.NextPageToken == "" {
			break
		}
		nextPageToken = allVersionsResponse.NextPageToken
	}
	return versions, nil
}

// DisableSecretVersion disables a version of a secret
// expects versionPath in "projects/*/secrets/*/versions/*" format
func (sm *Service) DisableSecretVersion(versionPath string) (*gcpsecretmanager.SecretVersion, error) {
	disableRequest := gcpsecretmanager.DisableSecretVersionRequest{}
	return sm.VersionService.Disable(versionPath, &disableRequest).Do()
}

// DestroySecretVersion destroys a version of a secret
// expects versionPath in "projects/*/secrets/*/versions/*" format
func (sm *Service) DestroySecretVersion(versionPath string) (*gcpsecretmanager.SecretVersion, error) {
	destroyRequest := gcpsecretmanager.DestroySecretVersionRequest{}
	return sm.VersionService.Destroy(versionPath, &destroyRequest).Do()
}

// GetLatestSecretVersionData retrieves payload of a latest secret version
// expects secretPath in "projects/*/secrets/*" format
func (sm *Service) GetLatestSecretVersionData(secretPath string) (string, error) {
	latestVersion, err := sm.GetLatestSecretVersion(secretPath)
	if err != nil {
		return "", err
	}
	latestVersionPath := latestVersion.Name
	secretVersionCall := sm.VersionService.Access(latestVersionPath)
	secretVersion, err := secretVersionCall.Do()
	if err != nil {
		return "", err
	}
	decodedSecretDataString, err := base64.StdEncoding.DecodeString(secretVersion.Payload.Data)
	return string(decodedSecretDataString), err
}

// GetSecretVersionData retrieves payload of a secret version
// expects secretPath in "projects/*/secrets/*/versions/*" format
func (sm *Service) GetSecretVersionData(secretPath string) (string, error) {
	secretVersionCall := sm.VersionService.Access(secretPath)
	secretVersion, err := secretVersionCall.Do()
	if err != nil {
		return "", err
	}
	decodedSecretDataString, err := base64.StdEncoding.DecodeString(secretVersion.Payload.Data)
	return string(decodedSecretDataString), err
}

// GetAllSecrets gets all or all filtered secrets from Secret Manager.
func (sm *Service) GetAllSecrets(projectPath string, filter string) ([]*gcpsecretmanager.Secret, error) {
	var secrets []*gcpsecretmanager.Secret
	nextPageToken := ""

	for {
		secretListCall := sm.Service.Projects.Secrets.List(projectPath)
		if filter != "" {
			secretListCall = secretListCall.Filter(filter)
		}
		if nextPageToken != "" {
			secretListCall.PageToken(nextPageToken)
		}

		secretList, err := secretListCall.Do()
		if err != nil {
			return nil, err
		}
		secrets = append(secrets, secretList.Secrets...)

		if secretList.NextPageToken == "" {
			break
		}
		nextPageToken = secretList.NextPageToken
	}

	return secrets, nil
}
# (2025-03-04)