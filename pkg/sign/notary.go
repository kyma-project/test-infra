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
	tagDelim          = ":"
	regRepoDelimiter  = "/"
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

// SimpleImage is a basic implementation of the Image interface
type SimpleImage struct {
	ManifestData Manifest
}

// Manifest returns the manifest data for the image
func (si *SimpleImage) Manifest() (*Manifest, error) {
	return &si.ManifestData, nil
}

// GetImage fetches the image manifest from a container registry
func GetImage(ref Reference) (Image, error) {
	r, ok := ref.(name.Reference)
	if !ok {
		return nil, fmt.Errorf("invalid reference type")
	}

	// Fetch the image from the registry
	img, err := remote.Image(r)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image: %w", err)
	}

	// Extract manifest from the image
	manifest, err := img.Manifest()
	if err != nil {
		return nil, fmt.Errorf("failed to get manifest: %w", err)
	}

	// Return the image, which implements the Image interface
	return &SimpleImage{
		ManifestData: Manifest{
			Config: struct {
				Digest struct {
					Hex string
				}
				Size int64
			}{
				Digest: struct {
					Hex string
				}{
					Hex: manifest.Config.Digest.Hex,
				},
				Size: manifest.Config.Size,
			},
		},
	}, nil
}

// SimpleReference is a basic implementation of the Reference interface
type SimpleReference struct {
	Image string
	Tag   string
}

func ParseReference(image string) (Reference, error) {
	ref, err := name.ParseReference(image)
	if err != nil {
		return nil, fmt.Errorf("failed to parse image reference: %w", err)
	}

	return ref, nil
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
func (ns *NotarySigner) Sign(images []string) error {
	sImg := strings.Join(images, ", ")

	// Build signing requests for the given images
	signingRequests, err := ns.buildSigningRequest(images)
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
	}

	// Initialize the HTTP client if not already initialized
	if ns.c == nil {
		ns.c = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
			Timeout: ns.retryTimeout,
		}
	} else {
		// Update the TLSConfig in the existing Transport if possible
		if transport, ok := ns.c.Transport.(*http.Transport); ok {
			// Set the TLSConfig
			transport.TLSClientConfig = tlsConfig
		} else {
			// Cannot set TLSConfig, perhaps it's a mock transport
			// In tests, this is acceptable
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

	// Exponential backoff settings
	backoff := time.Second // Start with 1 second

	for retries > 0 {
		fmt.Printf("Trying to sign %s. %v retries remaining...\n", sImg, retries)

		// Send the HTTP request
		resp, err := ns.c.Do(req)
		if err != nil {
			fmt.Printf("Request failed with error: %v. Retrying after %v...\n", err, backoff)
			retries--
			time.Sleep(backoff) // Wait for the backoff duration
			backoff *= 2        // Exponential backoff (double the time each retry)
			continue            // Retry on failure
		}
		status = resp.Status
		defer resp.Body.Close()

		// Read the response body
		respMsg, err = io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		// Check response status
		switch resp.StatusCode {
		case http.StatusAccepted:
			fmt.Printf("Successfully signed images %s!\n", sImg)
			return nil
		case http.StatusUnauthorized, http.StatusForbidden, http.StatusBadRequest, http.StatusUnsupportedMediaType:
			return fmt.Errorf("failed to sign images: %w", ErrBadResponse{status: status, message: string(respMsg)})
		}

		retries--
	}

	return fmt.Errorf("failed to sign images after retries: %w", ErrBadResponse{status: status, message: string(respMsg)})
}

// NewSigner creates a new NotarySigner based on the configuration.
func (nc NotaryConfig) NewSigner() (Signer, error) {
	ns := NotarySigner{
		retryTimeout:       10 * time.Second,
		url:                "https://signing-manage.repositories.cloud.sap/trusted-collections/publish",
		ParseReferenceFunc: ParseReference,
		GetImageFunc:       GetImage,
	}

	ns.BuildPayloadFunc = ns.buildPayload
	ns.DecodeCertFunc = ns.signifySecret.DecodeCertAndKey

	// Set retry timeout if configured
	if nc.RetryTimeout > 0 {
		ns.retryTimeout = nc.RetryTimeout
	}

	// Load secret if provided
	if nc.Secret != nil {
		switch nc.Secret.Type {
		case "signify":
			// Use injected ReadFileFunc or default to os.ReadFile
			readFileFunc := nc.ReadFileFunc
			if readFileFunc == nil {
				readFileFunc = os.ReadFile
			}
			f, err := readFileFunc(nc.Secret.Path)
			if err != nil {
				return nil, fmt.Errorf("failed to read secret file: %w", err)
			}

			var s SignifySecret
			err = json.Unmarshal(f, &s)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal signify secret: %w", err)
			}

			// Ensure signifySecret is properly initialized
			ns.signifySecret = s

		default:
			return nil, fmt.Errorf("unsupported secret type: %s", nc.Secret.Type)
		}
	} else {
		return nil, fmt.Errorf("no secret configuration provided")
	}

	// Only create a new HTTP client if one isn't provided (for example, in tests)
	if ns.c == nil {
		ns.c = &http.Client{
			Timeout: nc.Timeout,
		}
	}

	return &ns, nil
}
