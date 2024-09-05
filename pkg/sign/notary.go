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
	"go.step.sm/crypto/pemutil"
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
	c             http.Client
	url           string
	retryTimeout  time.Duration
	signifySecret SignifySecret
}

// DecodeCertAndKey loads the certificate and private key using smallstep/crypto and returns tls.Certificate
func (ss *SignifySecret) DecodeCertAndKey() (tls.Certificate, error) {
	// Parse the certificate from base64 encoded PEM data
	certData, err := base64.StdEncoding.DecodeString(ss.CertficateData)
	cert, err := pemutil.ParseCertificate(certData)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Parse the private key, handling password if necessary
	keyData, err := base64.StdEncoding.DecodeString(ss.PrivateKeyData)
	key, err := pemutil.ParseKey(keyData, pemutil.WithPassword([]byte(ss.KeyPassword)))
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Return the tls.Certificate that can be used for mTLS
	return tls.Certificate{
		Certificate: [][]byte{cert.Raw}, // The raw DER bytes of the certificate
		PrivateKey:  key,
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

// Sign makes an HTTP request to sign the images using the Notary server
func (ns NotarySigner) Sign(images []string) error {
	sImg := strings.Join(images, ", ")
	sr, err := ns.buildSigningRequest(images)
	if err != nil {
		return fmt.Errorf("build sign request: %w", err)
	}

	// Prepare the payload to send in the request
	payload := map[string]interface{}{
		"trustedCollections": []map[string]interface{}{
			{
				"gun": sr[0].NotaryGun,
				"targets": []map[string]interface{}{
					{
						"name":     sr[0].Version,
						"byteSize": sr[0].ByteSize,
						"digest":   sr[0].SHA256,
					},
				},
			},
		},
	}

	// Marshal the payload to JSON
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Decode the certificate and key using the smallstep utility
	cert, err := ns.signifySecret.DecodeCertAndKey()
	if err != nil {
		return fmt.Errorf("failed to load certificate and key: %w", err)
	}

	// Configure TLS with the decoded certificate and private key
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS10,
		MaxVersion:   tls.VersionTLS13, // or whichever version is required
	}

	// Create an HTTP client with TLS support
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: 3000 * time.Second, // Increase the timeout as necessary
	}

	// Create a new POST request with the signing payload
	req, err := http.NewRequest("POST", ns.url, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")

	retries := 5
	var respMsg []byte

	w := time.NewTicker(ns.retryTimeout)
	defer w.Stop()

	for retries > 0 {
		fmt.Printf("Trying to sign %s. %d retries remaining...\n", sImg, retries)
		<-w.C

		// Send the HTTP request
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send request: %w", err)
		}
		defer resp.Body.Close()

		// Read the response body
		respMsg, err = io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		// Check if the request was successful
		if resp.StatusCode == http.StatusAccepted {
			fmt.Println("Successfully signed images")
			return nil
		}

		retries--
	}

	return fmt.Errorf("failed to sign images: %s", string(respMsg))
}

func (nc NotaryConfig) NewSigner() (*NotarySigner, error) {
	var ns NotarySigner

	if nc.Secret != nil {
		f, err := os.ReadFile(nc.Secret.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to read secret file: %w", err)
		}

		switch nc.Secret.Type {
		case "signify":
			var s SignifySecret
			err := json.Unmarshal(f, &s)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal signify secret: %w", err)
			}

			ns.signifySecret = s

		default:
			return nil, fmt.Errorf("unsupported secret type: %s", nc.Secret.Type)
		}
	}

	ns.retryTimeout = 10 * time.Second
	if nc.RetryTimeout > 0 {
		ns.retryTimeout = nc.RetryTimeout
	}

	ns.url = "https://signing-manage-stage.repositories.cloud.sap/trusted-collections/publish"

	return &ns, nil
}
