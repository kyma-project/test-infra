package secretmanager

import (
	"context"
	"encoding/base64"
	"fmt"

	"google.golang.org/api/option"
	gcpsecretmanager "google.golang.org/api/secretmanager/v1"
)

func NewService(ctx context.Context, serviceAccountGCP string) (*Service, error) {
	secretManagerClient, err := gcpsecretmanager.NewService(ctx, option.WithCredentialsFile(serviceAccountGCP))
	if err != nil {
		return nil, fmt.Errorf("failed to create google Secret Manager client, got error: %w", err)
	}
	return &Service{Service: secretManagerClient}, nil
}

func (sm *Service) AddSecretVersion(secretPath string, secretData []byte) (*gcpsecretmanager.SecretVersion, error) {
	newVersionRequest := gcpsecretmanager.AddSecretVersionRequest{Payload: &gcpsecretmanager.SecretPayload{Data: base64.StdEncoding.EncodeToString(secretData)}}
	newVersionCall := sm.Projects.Secrets.AddVersion(secretPath, &newVersionRequest)
	secretVersion, err := newVersionCall.Do()
	return secretVersion, err
}

func (sm *Service) ListSecretVersions(secretPath string) (*gcpsecretmanager.ListSecretVersionsResponse, error) {
	secretVersionsCall := sm.Projects.Secrets.Versions.List(secretPath)
	secretVersions, err := secretVersionsCall.Do()
	return secretVersions, err
}

func (sm *Service) GetSecretVersion(secretPath string) (*gcpsecretmanager.SecretVersion, error) {
	secretVersionsCall := sm.Projects.Secrets.Versions.Get(secretPath)
	secretVersion, err := secretVersionsCall.Do()
	return secretVersion, err
}
