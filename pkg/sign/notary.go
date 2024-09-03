package sign

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

const (
	TypeNotaryBackend = "notary"
)

type ErrBadResponse struct {
	status  string
	message string
}

type ErrAuthServiceNotSupported struct {
	Service string
}

func (e ErrAuthServiceNotSupported) Error() string {
	return fmt.Sprintf("'%s' auth service not supported", e.Service)
}

func (e ErrBadResponse) Error() string {
	return fmt.Sprintf("bad response from service: %s, %s", e.status, e.message)
}

type NotaryConfig struct {
	// Set URL to Notary server signing endpoint
	Endpoint string `yaml:"endpoint" json:"endpoint"`
	// SecretPath contains path to the authentication credentials used for specific notary server
	Secret *AuthSecretConfig `yaml:"secret,omitempty" json:"secret,omitempty"`
	// Time after connection to notary server should time out
	Timeout time.Duration `yaml:"timeout" json:"timeout"`
	// RetryTimeout is time between each signing request to notary in case something fails
	// Default is 10 seconds
	RetryTimeout time.Duration `yaml:"retry-timeout" json:"retry-timeout"`

	CertFile   string `yaml:"certFile" json:"certFile"`
	KeyFile    string `yaml:"keyFile" json:"keyFile"`
	Passphrase string `yaml:"passphrase" json:"passphrase"`
}

// AuthSecretConfig contains auth information for notary server
type AuthSecretConfig struct {
	// Path if path to file that contains secret credentials
	Path string `yaml:"path" json:"path"`
	// Type contains credential type, based on which the service will configure signing
	Type string `yaml:"type" json:"type"`
}

// SignifySecret contains configuration of secret that is used to connect to SAP signify service
type SignifySecret struct {
	// URL to the SAP signify endpoint
	Endpoint string `yaml:"endpoint" json:"endpoint"`
	// Actual secret data that will be pushed on sign request. Assumed its in JSON format
	Payload string `yaml:"payload" json:"payload"`
}

// SigningRequest contains information about all images with tags to sign using Notary
type SigningRequest struct {
	// Global unique name, e.g. full image name with registry URL
	NotaryGun string `json:"notaryGun"`
	// SHA sum of manifest.json
	SHA256 string `json:"sha256"`
	// size of manifest.json
	ByteSize int64 `json:"byteSize"`
	// Image tag
	Version string `json:"version"`
}

// AuthFunc is a small middleware function that intercepts http.Request with specific authentication method
type AuthFunc func(r *http.Request) *http.Request

// NotarySigner is a struct that implements sign.Signer interface
// Takes care of signing requests to Notary server
type NotarySigner struct {
	client       *http.Client
	url          string
	retryTimeout time.Duration
}

// AuthToken mutates created request, so it contains bearer token of authorized user
// Serves as middleware function before sending request
func AuthToken(token string) AuthFunc {
	return func(r *http.Request) *http.Request {
		r.Header.Add("Authorization", "Token "+token)
		return r
	}
}

func (nc NotaryConfig) LoadCertFiles() (*tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(nc.CertFile, nc.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load cert or key file: %w", err)
	}
	return &cert, nil
}

// SignifyAuth fetches the JWT token from provided API endpoint in configuration file
// then returns a function that populates this JWT token as Bearer in the request
func SignifyAuth(s SignifySecret) (AuthFunc, error) {
	endpoint := s.Endpoint
	payload := s.Payload
	resp, err := http.Post(endpoint, "application/json", strings.NewReader(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var t struct {
		AccessToken struct {
			Token string `json:"token"`
		} `json:"access_token"`
	}

	rBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, ErrBadResponse{status: resp.Status, message: string(rBody)}
	}
	if err := json.Unmarshal(rBody, &t); err != nil {
		return nil, err
	}

	token := t.AccessToken.Token

	return func(r *http.Request) *http.Request {
		r.Header.Add("Authorization", "Bearer "+token)
		return r
	}, nil
}

func (ns NotarySigner) buildSigningRequest(images []string) ([]SigningRequest, error) {
	var sr []SigningRequest
	for _, i := range images {
		var base, tag string
		// Split on ":"
		parts := strings.Split(i, tagDelim)
		// Verify that we aren't confusing a tag for a hostname w/ port for the purposes of weak validation.
		if len(parts) > 1 && !strings.Contains(parts[len(parts)-1], regRepoDelimiter) {
			base = strings.Join(parts[:len(parts)-1], tagDelim)
			tag = parts[len(parts)-1]
		}
		ref, err := name.ParseReference(i)
		if err != nil {
			return nil, fmt.Errorf("ref parse: %w", err)
		}
		i, err := remote.Image(ref)
		if err != nil {
			return nil, fmt.Errorf("get image: %w", err)
		}
		m, err := i.Manifest()
		if err != nil {
			return nil, fmt.Errorf("image manifest: %w", err)
		}
		sha := m.Config.Digest.Hex
		size := m.Config.Size
		sr = append(sr, SigningRequest{NotaryGun: base, Version: tag, ByteSize: size, SHA256: sha})
	}
	return sr, nil
}

func (ns NotarySigner) Sign(images []string) error {
	sr, err := ns.buildSigningRequest(images)
	if err != nil {
		return fmt.Errorf("failed to build signing request: %w", err)
	}

	payload, err := json.Marshal(map[string]interface{}{
		"trustedCollections": sr,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal signing request: %w", err)
	}

	// Create a new HTTP request for the signing operation
	req, err := http.NewRequest("POST", ns.url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create new HTTP request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")

	retries := 3 // Number of retry attempts
	var lastErr error

	// Retry loop for handling transient errors during the signing process
	for retries > 0 {
		// Execute the HTTP request
		resp, err := ns.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("HTTP request failed: %w", err)
			retries--
			if retries > 0 {
				fmt.Printf("Request failed, retrying... (%d retries left)\n", retries)
				time.Sleep(ns.retryTimeout) // Wait before retrying
				continue
			}
			return lastErr
		}
		defer resp.Body.Close()

		// Check if the response status code indicates success
		if resp.StatusCode != http.StatusAccepted {
			respBody, _ := io.ReadAll(resp.Body)
			lastErr = fmt.Errorf("signing failed, status: %s, response: %s", resp.Status, string(respBody))
			retries--
			if retries > 0 {
				fmt.Printf("Received non-accepted status %d, retrying... (%d retries left)\n", resp.StatusCode, retries)
				time.Sleep(ns.retryTimeout)
				continue
			}
			return lastErr
		}

		// If the request is successful, log the outcome and exit the loop
		fmt.Println("Images signed successfully.")
		return nil
	}

	// Return the last encountered error after all retry attempts are exhausted
	return lastErr
}

func (nc NotaryConfig) NewSigner() (Signer, error) {
	cert, err := nc.LoadCertFiles()
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*cert},
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: nc.Timeout,
	}

	return &NotarySigner{
		client:       client,
		url:          nc.Endpoint,
		retryTimeout: nc.RetryTimeout,
	}, nil
}
