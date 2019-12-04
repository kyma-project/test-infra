package secrets

import (
	"context"
	"fmt"

	kms "cloud.google.com/go/kms/apiv1"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"

	"github.com/kyma-project/test-infra/development/prow-installer/pkg/storage"
)

// Option wrapper for relevant Options for the client
type Option struct {
	Prefix     string // storage prefix
	ProjectID  string // GCP project ID
	LocationID string // location of the key rings
	KmsRing    string // kms keyring name
	KmsKey     string // kms key of keyring
	Bucket     string // storage bucket name
}

// Client wrapper for KMS and GCS secret storage
type Client struct {
	Option
	ctx context.Context
}

// New returns a new Client, wrapping kms and gcs for storing/reading encrypted secrets from GCP
func New(ctx context.Context, opts Option) (*Client, error) {
	if opts.ProjectID == "" {
		return nil, fmt.Errorf("ProjectID is required to initialize a client")
	}
	if opts.LocationID == "" {
		return nil, fmt.Errorf("LocationID is required to initialize a client")
	}
	if opts.KmsRing == "" {
		return nil, fmt.Errorf("KmsRing is required to initialize a client")
	}
	if opts.KmsKey == "" {
		return nil, fmt.Errorf("KmsKey is required to initialize a client")
	}
	if opts.Bucket == "" {
		return nil, fmt.Errorf("Bucket is required to initialize a client")
	}
	return &Client{Option: opts, ctx: ctx}, nil
}

// StoreSecret encrypts and stores a secret value using KMS and GCS API.
func (sc *Client) StoreSecret(plaintext []byte, storageObject string) error {
	data, err := sc.encrypt(plaintext)
	if err != nil {
		return fmt.Errorf("Encrypting secret failed: %w", err)
	}

	client, err := storage.New(sc.ctx, storage.Option{Prefix: sc.Prefix, ProjectID: sc.ProjectID, LocationID: sc.LocationID})
	if err != nil {
		return fmt.Errorf("Could not create GCS Storage Client: %v", err)

	}

	if err := client.Write(data, sc.Bucket, sc.prefixedName(storageObject)); err != nil {
		return fmt.Errorf("Storing secret failed: %w", err)
	}
	return nil
}

// ReadSecret reads a secret value and decrypts it using KMS and GCS API.
func (sc *Client) ReadSecret(storageObject string) ([]byte, error) {
	client, err := storage.New(sc.ctx, storage.Option{Prefix: sc.Prefix, ProjectID: sc.ProjectID, LocationID: sc.LocationID})
	if err != nil {
		return nil, fmt.Errorf("Could not create GCS Storage Client: %v", err)
	}

	data, err := client.Read(sc.Bucket, sc.prefixedName(storageObject))
	if err != nil {
		return nil, fmt.Errorf("Reading secret failed: %w", err)
	}

	plain, err := sc.decrypt(data)
	if err != nil {
		return nil, fmt.Errorf("Decrypting secret failed: %w", err)
	}

	return plain, nil
}

func (sc *Client) encrypt(plaintext []byte) ([]byte, error) {
	client, err := kms.NewKeyManagementClient(sc.ctx)
	if err != nil {
		return nil, fmt.Errorf("Initialising KMS client failed: %w", err)
	}

	req := &kmspb.EncryptRequest{
		Name:      sc.getKmsKeyPath(),
		Plaintext: plaintext,
	}
	resp, err := client.Encrypt(sc.ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Encrypting secret failed: %w", err)
	}

	return resp.Ciphertext, nil
}

func (sc *Client) decrypt(ciphertext []byte) ([]byte, error) {
	client, err := kms.NewKeyManagementClient(sc.ctx)
	if err != nil {
		return nil, fmt.Errorf("Initialising KMS client failed: %w", err)
	}

	// Build the request.
	req := &kmspb.DecryptRequest{
		Name:       sc.getKmsKeyPath(),
		Ciphertext: ciphertext,
	}
	// Call the API.
	resp, err := client.Decrypt(sc.ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Decrypting secret failed: %w", err)
	}
	return resp.Plaintext, nil
}

func (sc *Client) prefixedName(name string) string {
	if sc.Prefix != "" {
		return fmt.Sprintf("%s-%s", sc.Prefix, name)
	}
	return name
}

func (sc *Client) getKmsKeyPath() string {
	return fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s", sc.ProjectID, sc.LocationID, sc.KmsRing, sc.KmsKey)
}
