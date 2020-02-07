package secrets

import (
	kms "cloud.google.com/go/kms/apiv1"
	"context"
	"fmt"
	"google.golang.org/api/option"
)

//go:generate mockery -name=API -output=automock -outpkg=automock -case=underscore

// Option wrapper for relevant Options for the client
type Option struct {
	ProjectID      string
	LocationID     string
	KmsRing        string
	KmsKey         string
	ServiceAccount string // filename of the serviceaccount to use
}

// Client wrapper for KMS and GCS secret storage
type Client struct {
	Option
	api API
}

// API provides a mockable interface for the GCP api. Find the implementation of the GCP wrapped API in wrapped.go
type API interface {
	Encrypt(ctx context.Context, plaintext []byte) ([]byte, error)
	Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error)
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

// New returns a new Client, wrapping kms and gcs for storing/reading encrypted secrets from GCP
func New(opts Option, api API) (*Client, error) {
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
	if opts.ServiceAccount == "" {
		return nil, fmt.Errorf("ServiceAccount is required to initialize a client")
	}
	if api == nil {
		return nil, fmt.Errorf("Can't create client without api")
	}

	return &Client{Option: opts, api: api}, nil
}

// Encrypt attempts to encrypt a secret with kms
func (sc *Client) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return nil, fmt.Errorf("cannot encrypt zero value")
	}
	return sc.api.Encrypt(ctx, plaintext)
}

// Decrypt attempts to decrypt a secret with kms
func (sc *Client) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) == 0 {
		return nil, fmt.Errorf("cannot decrypt zero value")
	}
	return sc.api.Decrypt(ctx, ciphertext)
}

// WithKmsRing modifies option to have a kms ring
func (o Option) WithKmsRing(ring string) Option {
	o.KmsRing = ring
	return o
}

// WithKmsKey modifies option to have a kms key
func (o Option) WithKmsKey(key string) Option {
	o.KmsKey = key
	return o
}

// WithProjectID modifies option to have a project id
func (o Option) WithProjectID(pid string) Option {
	o.ProjectID = pid
	return o
}

// WithLocationID modifies option to have a zone id
func (o Option) WithLocationID(l string) Option {
	o.LocationID = l
	return o
}

// WithServiceAccount modifies option to have a service account
func (o Option) WithServiceAccount(sa string) Option {
	o.ServiceAccount = sa
	return o
}
