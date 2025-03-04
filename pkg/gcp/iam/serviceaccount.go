package iam

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"

	gcpiam "google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
)

// ServiceAccountJSON stores Service Account authentication data
type ServiceAccountJSON struct {
	Type             string `json:"type"`
	ProjectID        string `json:"project_id"`
	PrivateKeyID     string `json:"private_key_id"`
	PrivateKey       string `json:"private_key"`
	ClientEmail      string `json:"client_email"`
	ClientID         string `json:"client_id"`
	AuthURL          string `json:"auth_uri"`
	TokenURI         string `json:"token_uri"`
	AuthProviderCert string `json:"auth_provider_x509_cert_url"`
	ClientCert       string `json:"client_x509_cert_url"`
}
type Service struct {
	*gcpiam.Service
}

func NewService(ctx context.Context, options ...option.ClientOption) (*Service, error) {
	iamClient, err := gcpiam.NewService(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create google Secret Manager client, got error: %w", err)
	}

	return &Service{Service: iamClient}, nil
}

func (s *Service) CreateNewServiceAccountKey(saPath string) ([]byte, error) {
	createKeyRequest := gcpiam.CreateServiceAccountKeyRequest{}
	newKeyCall := s.Projects.ServiceAccounts.Keys.Create(saPath, &createKeyRequest)
	newKey, err := newKeyCall.Do()
	if err != nil {
		return []byte{}, err
	}

	log.Printf("Decoding new key data for %s", saPath)
	newKeyBytes, err := base64.StdEncoding.DecodeString(newKey.PrivateKeyData)
	if err != nil {
		return []byte{}, err
	}
	return newKeyBytes, nil
}

func (s *Service) DeleteKey(serviceAccountKeyPath string) error {
	keyVersionCall := s.Projects.ServiceAccounts.Keys.Delete(serviceAccountKeyPath)
	_, err := keyVersionCall.Do()
	return err
}
# (2025-03-04)