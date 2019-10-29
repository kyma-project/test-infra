package secrets

import (
	"context"
	"fmt"
	"io/ioutil"

	kms "cloud.google.com/go/kms/apiv1"
	gcs "cloud.google.com/go/storage"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

// Option wrapper for relevant Options for the client
type Option struct {
	Prefix     string // storage prefix
	ProjectID  string // GCP project ID
	LocationID string // location of the key rings
}

// Client wrapper for KMS and GCS secret storage
type Client struct {
	Option
	ctx context.Context
}

// New returns a new Client, wrapping kms and gcs for storing/reading encrypted secrets from GCP
func New(ctx context.Context, opts Option) (*Client, error) {
	if opts.Prefix == "" {
		return nil, fmt.Errorf("Prefix is required to initialize a client")
	}
	if opts.ProjectID == "" {
		return nil, fmt.Errorf("ProjectID is required to initialize a client")
	}
	if opts.LocationID == "" {
		return nil, fmt.Errorf("LocationID is required to initialize a client")
	}
	return &Client{Option: opts, ctx: ctx}, nil
}

// StoreSecret encrypts and stores a secret value using KMS and GCS API.
func (sc *Client) StoreSecret(secret []byte, bucket, keyName, secretName string) error {
	if err := sc.write(secret, bucket, keyName, secretName); err != nil {
		return fmt.Errorf("Storing secret failed: %v", err)
	}
	return nil
}

// ReadSecret reads a secret value and decrypts it using KMS and GCS API.
func (sc *Client) ReadSecret(bucket, keyName, secretName string) ([]byte, error) {
	data, err := sc.read(bucket, secretName)
	if err != nil {
		return nil, err
	}

	plain, err := sc.decrypt(keyName, data)
	if err != nil {
		return nil, err
	}

	return plain, nil
}

func (sc *Client) decrypt(keyName string, ciphertext []byte) ([]byte, error) {
	client, err := kms.NewKeyManagementClient(sc.ctx)
	if err != nil {
		return nil, err
	}

	// Build the request.
	req := &kmspb.DecryptRequest{
		Name:       keyName,
		Ciphertext: ciphertext,
	}
	// Call the API.
	resp, err := client.Decrypt(sc.ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Decrypt: %v", err)
	}
	return resp.Plaintext, nil
}

func (sc *Client) read(bucket, secretName string) ([]byte, error) {
	client, err := gcs.NewClient(sc.ctx)
	if err != nil {
		return nil, err
	}
	rc, err := client.Bucket(bucket).Object(secretName).NewReader(sc.ctx)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (sc *Client) write(data []byte, bucket, keyName, secretName string) error {
	client, err := gcs.NewClient(sc.ctx)
	if err != nil {
		return err
	}
	obj := client.Bucket(bucket).Object(secretName)
	// Encrypt the object's contents.
	wc := obj.NewWriter(sc.ctx)
	wc.KMSKeyName = keyName
	if _, err := wc.Write(data); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}
	return nil
}
