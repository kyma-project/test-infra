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

	// ReadFileFunc allows injecting a custom file reading function, defaults to os.ReadFile.
	ReadFileFunc func(string) ([]byte, error)
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
	// Certificate data encoded in base64
	CertificateData string `json:"certData"`
	// Private key data encoded in base64
	PrivateKeyData string `json:"privateKeyData"`
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

// Target represents the target data for signing
type Target struct {
	Name     string `json:"name"`
	ByteSize int64  `json:"byteSize"`
	Digest   string `json:"digest"`
}

// TrustedCollection represents a trusted collection for a specific image
type TrustedCollection struct {
	GUN     string   `json:"gun"`
	Targets []Target `json:"targets"`
}

// SigningPayload represents the overall payload structure for the signing request
type SigningPayload struct {
	TrustedCollections []TrustedCollection `json:"trustedCollections"`
}

// NotarySigner is a struct that implements sign.Signer interface
// Takes care of signing requests to Notary server
type NotarySigner struct {
	c                   *http.Client
	url                 string
	retryTimeout        time.Duration
	signifySecret       SignifySecret
	ParseReferenceFunc  func(image string) (Reference, error)
	GetImageFunc        func(ref Reference) (Image, error)
	DecodeCertFunc      func() (tls.Certificate, error)
	BuildSigningReqFunc func([]string) ([]SigningRequest, error)
	BuildPayloadFunc    func([]SigningRequest) (SigningPayload, error)
}

type Reference interface{}
type Image interface {
	Manifest() (*Manifest, error)
}

type Manifest struct {
	Config struct {
		Digest struct {
			Hex string
		}
		Size int64
	}
}

// DecodeCertAndKey loads the certificate and decrypted private key from base64-encoded strings in SignifySecret
func (ss *SignifySecret) DecodeCertAndKey() (tls.Certificate, error) {
	// Decode the base64-encoded certificate
	certData, err := base64.StdEncoding.DecodeString(ss.CertificateData)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to decode certificate: %w", err)
	}

	// Decode the base64-encoded private key
	keyData, err := base64.StdEncoding.DecodeString(ss.PrivateKeyData)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to decode private key: %w", err)
	}

	// Load the certificate and key as a TLS certificate pair
	cert, err := tls.X509KeyPair(certData, keyData)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("unable to load certificate or key: %w", err)
	}

	return cert, nil
}

// buildSigningRequest prepares the signing requests for the given images
func (ns NotarySigner) buildSigningRequest(images []string) ([]SigningRequest, error) {
	var signingRequests []SigningRequest

	for _, i := range images {
		var base, tag string

		// Split on ":" to separate base from tag
		parts := strings.Split(i, tagDelim)
		if len(parts) > 1 && !strings.Contains(parts[len(parts)-1], regRepoDelimiter) {
			base = strings.Join(parts[:len(parts)-1], tagDelim)
			tag = parts[len(parts)-1]
		}

		ref, err := ns.ParseReferenceFunc(i)
		if err != nil {
			return nil, fmt.Errorf("ref parse: %w", err)
		}
		img, err := ns.GetImageFunc(ref)
		if err != nil {
			return nil, fmt.Errorf("get image: %w", err)
		}
		manifest, err := img.Manifest()
		if err != nil {
			return nil, fmt.Errorf("failed getting image manifest: %w", err)
		}

		signingRequests = append(signingRequests, SigningRequest{
			NotaryGun: base,
			SHA256:    manifest.Config.Digest.Hex,
			ByteSize:  manifest.Config.Size,
			Version:   tag,
		})
	}

	return signingRequests, nil
}

// buildPayload creates the payload for the signing request from a list of SigningRequests
func (ns NotarySigner) buildPayload(sr []SigningRequest) (SigningPayload, error) {
	var trustedCollections []TrustedCollection

	// Loop through all signing requests and create separate entries for each GUN
	for _, req := range sr {
		target := Target{
			Name:     req.Version,
			ByteSize: req.ByteSize,
			Digest:   req.SHA256,
		}

		// Each image gets its own trusted collection based on its GUN
		trustedCollection := TrustedCollection{
			GUN:     req.NotaryGun,
			Targets: []Target{target},
		}

		trustedCollections = append(trustedCollections, trustedCollection)
	}

	// Prepare the payload structure with multiple trustedCollections
	payload := SigningPayload{
		TrustedCollections: trustedCollections,
	}

	return payload, nil
}

// Sign makes an HTTP request to sign the images using the Notary server
func (ns NotarySigner) Sign(images []string) error {
	sImg := strings.Join(images, ", ")

	// Get the signing requests
	signingRequests, err := ns.BuildSigningReqFunc(images)
	if err != nil {
		return fmt.Errorf("build signing request: %w", err)
	}

	// Build the payload from signing requests
	payload, err := ns.BuildPayloadFunc(signingRequests)
	if err != nil {
		return fmt.Errorf("build payload: %w", err)
	}

	// Marshal the payload to JSON
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal signing request: %w", err)
	}

	// Decode the certificate and key from the signifySecret structure
	cert, err := ns.DecodeCertFunc()
	if err != nil {
		return fmt.Errorf("failed to load certificate and key: %w", err)
	}

	// Configure TLS with the decoded certificate and private key
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	// Ensure we are using the injected HTTP client with the mocked transport
	client := ns.c
	if client == nil {
		// Default to real client if not injected
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
			Timeout: ns.retryTimeout,
		}
	}

	// Create a new POST request with the signing payload
	req, err := http.NewRequest("POST", ns.url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	retries := 5
	var status string
	var respMsg []byte
	w := time.NewTicker(ns.retryTimeout)
	defer w.Stop()
	for retries > 0 {
		fmt.Printf("Trying to sign %s. %v retries remaining...\n", sImg, retries)
		<-w.C

		// Send the HTTP request
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send request: %w", err)
		}
		status = resp.Status
		defer resp.Body.Close()

		// Read the response body
		respMsg, err = io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}
		switch resp.StatusCode {
		case http.StatusAccepted:
			fmt.Printf("Successfully signed images %s!\n", sImg)
			return nil
		case http.StatusUnauthorized, http.StatusForbidden, http.StatusBadRequest, http.StatusUnsupportedMediaType:
			return fmt.Errorf("failed to sign images: %w", ErrBadResponse{status: status, message: string(respMsg)})
		}
		retries--
	}

	return fmt.Errorf("failed to sign images: %w", ErrBadResponse{status: status, message: string(respMsg)})
}

// NewSigner creates a new NotarySigner based on the configuration.
func (nc NotaryConfig) NewSigner() (Signer, error) {
	var ns NotarySigner

	// Ensure nc.Secret is not nil
	if nc.Secret != nil {
		// Check secret type before reading the file
		switch nc.Secret.Type {
		case "signify":
			// Use injected ReadFileFunc or default to os.ReadFile
			readFileFunc := nc.ReadFileFunc
			if readFileFunc == nil {
				readFileFunc = os.ReadFile
			}

			// Read the secret file
			f, err := readFileFunc(nc.Secret.Path)
			if err != nil {
				return nil, fmt.Errorf("failed to read secret file: %w", err)
			}

			// Unmarshal signify secret
			var s SignifySecret
			err = json.Unmarshal(f, &s)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal signify secret: %w", err)
			}

			ns.signifySecret = s

		default:
			return nil, fmt.Errorf("unsupported secret type: %s", nc.Secret.Type)
		}
	}

	// Initialize the HTTP client
	ns.c = &http.Client{}

	ns.retryTimeout = 10 * time.Second
	if nc.RetryTimeout > 0 {
		ns.retryTimeout = nc.RetryTimeout
	}
	ns.c.Timeout = nc.Timeout

	ns.url = "https://signing-manage-stage.repositories.cloud.sap/trusted-collections/publish"

	return &ns, nil
}
