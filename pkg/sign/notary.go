package sign

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
	CreatedAt      float64 `json:"createdAt"`
	TokenURL       string  `json:"tokenURL"`
	CertServiceURL string  `json:"certServiceURL"`
	ClientID       string  `json:"clientID"`
	CertficateData string  `json:"certData"`
	PrivateKeyData string  `json:"privateKeyData"`
	KeyPassword    string  `json:"password"`
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

// NotarySigner is a struct that implements sign.Signer interface
// Takes care of signing requests to Notary server
type NotarySigner struct {
	c            http.Client
	url          string
	retryTimeout time.Duration
}

func (ss *SignifySecret) DecodeCertAndKey() (tls.Certificate, error) {
	// Decode the certificate and private key from base64
	certData, err := base64.StdEncoding.DecodeString(ss.CertficateData)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to decode certificate data: %w", err)
	}

	keyData, err := base64.StdEncoding.DecodeString(ss.PrivateKeyData)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to decode private key data: %w", err)
	}

	// Load the certificate and key
	cert, err := tls.X509KeyPair(certData, keyData)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to load certificate and key: %w", err)
	}

	return cert, nil
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
		img, err := remote.Image(ref)
		if err != nil {
			return nil, fmt.Errorf("get image: %w", err)
		}
		m, err := img.Manifest()
		if err != nil {
			return nil, fmt.Errorf("image manifest: %w", err)
		}
		sha := m.Config.Digest.Hex
		size := m.Config.Size
		sr = append(sr, SigningRequest{
			NotaryGun: base,
			Version:   tag,
			ByteSize:  size,
			SHA256:    sha,
		})
	}
	return sr, nil
}

func (ns NotarySigner) Sign(images []string) error {
	// Build the signing request
	sImg := strings.Join(images, ", ")
	sr, err := ns.buildSigningRequest(images)
	if err != nil {
		return fmt.Errorf("build sign request: %w", err)
	}

	// Build the Signify API payload
	payload := map[string]interface{}{
		"trustedCollections": []map[string]interface{}{
			{
				"gun": sr[0].NotaryGun, // Example: "example.repo/image-project-2"
				"targets": []map[string]interface{}{
					{
						"name":     sr[0].Version,  // Example: "1.0.1"
						"byteSize": sr[0].ByteSize, // Size of the image's manifest
						"digest":   sr[0].SHA256,   // SHA-256 of the image's manifest
					},
				},
			},
		},
	}

	// Marshal the payload to JSON
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal signing request: %w", err)
	}

	// Create a new POST request with the signing payload
	req, err := http.NewRequest("POST", ns.url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	// Retry logic for sending the request
	retries := 5
	var respMsg []byte
	var status string
	w := time.NewTicker(ns.retryTimeout)
	defer w.Stop()

	for retries > 0 {
		fmt.Printf("Trying to sign %s. %d retries remaining...\n", sImg, retries)
		<-w.C

		// Send the HTTP request
		resp, err := ns.c.Do(req)
		if err != nil {
			return fmt.Errorf("notary request: %w", err)
		}
		defer resp.Body.Close()

		// Read the response body
		respMsg, err = io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("body read: %w", err)
		}
		status = resp.Status

		// Handle different response statuses
		switch resp.StatusCode {
		case http.StatusAccepted:
			fmt.Printf("Successfully signed images %s!\n", sImg)
			return nil
		case http.StatusUnauthorized, http.StatusForbidden, http.StatusBadRequest:
			return ErrBadResponse{status: status, message: string(respMsg)}
		}

		retries--
	}

	// After retries, return the error if signing fails
	fmt.Println("Reached all retries. Stopping.")
	return ErrBadResponse{status: status, message: string(respMsg)}
}

func (nc NotaryConfig) NewSigner() (*NotarySigner, error) {
	var ns NotarySigner

	// Load the secret file and initialize the signer
	if nc.Secret != nil {
		f, err := os.ReadFile(nc.Secret.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to read secret file: %w", err)
		}

		switch nc.Secret.Type {
		case "signify":
			// Unmarshal the YAML secret file into SignifySecret struct
			var s SignifySecret
			err := json.Unmarshal(f, &s)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal signify secret: %w", err)
			}

			// Decode certificate and key from base64
			cert, err := s.DecodeCertAndKey()
			if err != nil {
				return nil, fmt.Errorf("failed to decode cert and key: %w", err)
			}

			// Set up the TLS configuration with the decoded cert
			tlsConfig := &tls.Config{
				Certificates: []tls.Certificate{cert},
			}

			// Create an HTTP client with TLS config
			ns.c = http.Client{
				Transport: &http.Transport{
					TLSClientConfig: tlsConfig,
				},
				Timeout: nc.Timeout,
			}
		default:
			return nil, ErrAuthServiceNotSupported{Service: nc.Secret.Type}
		}
	}

	// Set retry timeout
	ns.retryTimeout = 10 * time.Second
	if nc.RetryTimeout > 0 {
		ns.retryTimeout = nc.RetryTimeout
	}

	// Set the Notary server URL
	ns.url = "https://signing-manage-stage.repositories.cloud.sap/trusted-collections/publish"

	return &ns, nil
}
