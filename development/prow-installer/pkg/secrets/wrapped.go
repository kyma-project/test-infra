package secrets

import (
	"context"
	"fmt"
	"google.golang.org/api/option"

	kms "cloud.google.com/go/kms/apiv1"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

// APIWrapper wraps the GCP api
type APIWrapper struct {
	ProjectID  string
	LocationID string
	KmsRing    string
	KmsKey     string
	KmsClient  *kms.KeyManagementClient
}

func NewClinet(ctx context.Context, opts Option, credentials string) (*Client, error) {
	kmsClient, err := kms.NewKeyManagementClient(ctx, option.WithCredentialsFile(credentials))
	if err != nil {
		return nil, fmt.Errorf("kms client create error %w", err)
	}

	api := &APIWrapper{
		ProjectID:  opts.ProjectID,
		LocationID: opts.LocationID,
		KmsRing:    opts.KmsRing,
		KmsKey:     opts.KmsKey,
		KmsClient:  kmsClient,
	}

	if client, err := New(opts, api); err != nil {
		return nil, fmt.Errorf("secrets client create error %w", err)
	} else {
		return client, nil
	}
}

// Encrypt calls the wrapped GCP api to encrypt a secret
func (caw *APIWrapper) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	req := &kmspb.EncryptRequest{
		Name:      caw.getKmsKeyPath(),
		Plaintext: plaintext,
	}
	resp, err := caw.KmsClient.Encrypt(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Encrypting secret failed: %w", err)
	}

	return resp.Ciphertext, nil
}

// Decrypt calls the wrapped GCP api to decrypt a secret
func (caw *APIWrapper) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	// Build the request.
	req := &kmspb.DecryptRequest{
		Name:       caw.getKmsKeyPath(),
		Ciphertext: ciphertext,
	}
	// Call the API.
	resp, err := caw.KmsClient.Decrypt(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Decrypting secret failed: %w", err)
	}
	return resp.Plaintext, nil
}

func (caw *APIWrapper) getKmsKeyPath() string {
	return fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s", caw.ProjectID, caw.LocationID, caw.KmsRing, caw.KmsKey)
}
