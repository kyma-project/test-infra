// Package sign provides functionality for signing container images using Notary v2.
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
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// ImageRepositoryInterface defines methods for parsing image references and fetching images.
type ImageRepositoryInterface interface {
	// ParseReference parses an image string into a ReferenceInterface.
	ParseReference(image string) (ReferenceInterface, error)
	// GetImage retrieves an image given its reference.
	GetImage(ref ReferenceInterface) (ImageInterface, error)
}

// ReferenceInterface abstracts the functionality of name.Reference.
type ReferenceInterface interface {
	// Name returns the full name of the reference.
	Name() string
	// String returns the string representation of the reference.
	String() string
	// GetRepositoryName returns the repository name from the reference.
	GetRepositoryName() string
	// GetTag returns the tag associated with the reference.
	GetTag() (string, error)
}

// ImageInterface abstracts the functionality of v1.Image.
type ImageInterface interface {
	// Manifest retrieves the manifest of the image.
	Manifest() (ManifestInterface, error)
}

// ManifestInterface abstracts the functionality of v1.Manifest.
type ManifestInterface interface {
	// GetConfigSize returns the size of the image config.
	GetConfigSize() int64
	// GetConfigDigest returns the digest of the image config.
	GetConfigDigest() string
}

// ImageService provides methods to parse image references and fetch images.
type ImageService struct{}

// ParseReference parses the image string into a ReferenceInterface.
func (is *ImageService) ParseReference(image string) (ReferenceInterface, error) {
	ref, err := name.ParseReference(image)
	if err != nil {
		return nil, fmt.Errorf("failed to parse image reference: %w", err)
	}
	return &ReferenceWrapper{Ref: ref}, nil
}

// GetImage fetches the image from the remote registry using the provided reference.
func (is *ImageService) GetImage(ref ReferenceInterface) (ImageInterface, error) {
	rw, ok := ref.(*ReferenceWrapper)
	if !ok {
		return nil, fmt.Errorf("unexpected reference type")
	}
	img, err := remote.Image(rw.Ref)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image: %w", err)
	}
	return &ImageWrapper{img: img}, nil
}

// ReferenceWrapper wraps a name.Reference to implement the ReferenceInterface.
type ReferenceWrapper struct {
	Ref name.Reference
}

// Name returns the full name of the reference.
func (rw *ReferenceWrapper) Name() string {
	return rw.Ref.Name()
}

// String returns the string representation of the reference.
func (rw *ReferenceWrapper) String() string {
	return rw.Ref.String()
}

// GetRepositoryName extracts and returns the repository name from the reference.
func (rw *ReferenceWrapper) GetRepositoryName() string {
	switch ref := rw.Ref.(type) {
	case name.Tag:
		return ref.Context().Name()
	case name.Digest:
		return ref.Context().Name()
	default:
		return ""
	}
}

// GetTag returns the tag associated with the reference.
func (rw *ReferenceWrapper) GetTag() (string, error) {
	switch ref := rw.Ref.(type) {
	case name.Tag:
		return ref.TagStr(), nil
	default:
		return "", fmt.Errorf("reference is not a tag")
	}
}

// ImageWrapper wraps a v1.Image to implement the ImageInterface.
type ImageWrapper struct {
	img v1.Image
}

// Manifest retrieves the manifest of the image.
func (iw *ImageWrapper) Manifest() (ManifestInterface, error) {
	manifest, err := iw.img.Manifest()
	if err != nil {
		return nil, err
	}
	return &ManifestWrapper{manifest: manifest}, nil
}

// ManifestWrapper wraps a v1.Manifest to implement the ManifestInterface.
type ManifestWrapper struct {
	manifest *v1.Manifest
}

// GetConfigSize returns the size of the image config.
func (mw *ManifestWrapper) GetConfigSize() int64 {
	return mw.manifest.Config.Size
}

// GetConfigDigest returns the digest of the image config.
func (mw *ManifestWrapper) GetConfigDigest() string {
	return mw.manifest.Config.Digest.String()
}

// PayloadBuilderInterface defines the method for constructing the signing payload.
type PayloadBuilderInterface interface {
	// BuildPayload constructs the signing payload for the provided images.
	BuildPayload(images []string) (SigningPayload, error)
}

// PayloadBuilder constructs the signing payload using an ImageRepositoryInterface.
type PayloadBuilder struct {
	ImageService ImageRepositoryInterface
}

// BuildPayload builds the signing payload for the given images.
func (pb *PayloadBuilder) BuildPayload(images []string) (SigningPayload, error) {
	var gunTargets []GUNTargets
	for _, image := range images {
		// Parse the image reference.
		ref, err := pb.ImageService.ParseReference(image)
		if err != nil {
			return SigningPayload{}, fmt.Errorf("failed to parse image reference: %w", err)
		}

		// Extract repository name and tag.
		base := ref.GetRepositoryName()
		tag, err := ref.GetTag()
		if err != nil {
			return SigningPayload{}, fmt.Errorf("failed to get tag from reference: %w", err)
		}

		// Fetch the image.
		img, err := pb.ImageService.GetImage(ref)
		if err != nil {
			return SigningPayload{}, fmt.Errorf("failed to fetch image: %w", err)
		}

		// Retrieve the image manifest.
		manifest, err := img.Manifest()
		if err != nil {
			return SigningPayload{}, fmt.Errorf("failed to get image manifest: %w", err)
		}

		// Build the target information.
		target := Target{
			Name:     tag,
			ByteSize: manifest.GetConfigSize(),
			Digest:   manifest.GetConfigDigest(),
		}

		// Build the GUN (Global Unique Name) targets.
		gunTarget := GUNTargets{
			GUN:     base,
			Targets: []Target{target},
		}
		gunTargets = append(gunTargets, gunTarget)
	}
	payload := SigningPayload{
		GunTargets: gunTargets,
	}
	return payload, nil
}

// TLSProviderInterface defines the method for obtaining TLS configuration.
type TLSProviderInterface interface {
	// GetTLSConfig returns a TLS configuration based on provided credentials.
	GetTLSConfig() (*tls.Config, error)
}

// TLSProvider provides TLS configurations using the provided TLS credentials.
type TLSProvider struct {
	Credentials TLSCredentials
}

// GetTLSConfig constructs a tls.Config using the stored TLS credentials.
func (tp *TLSProvider) GetTLSConfig() (*tls.Config, error) {
	certData, err := base64.StdEncoding.DecodeString(tp.Credentials.CertificateData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode certificate data: %w", err)
	}
	keyData, err := base64.StdEncoding.DecodeString(tp.Credentials.PrivateKeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key data: %w", err)
	}
	cert, err := tls.X509KeyPair(certData, keyData)
	if err != nil {
		return nil, fmt.Errorf("unable to load certificate and key: %w", err)
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	return tlsConfig, nil
}

// HTTPClientInterface defines methods for making HTTP requests and setting TLS configurations.
type HTTPClientInterface interface {
	// Do sends an HTTP request and returns an HTTP response.
	Do(req *http.Request) (*http.Response, error)
	// SetTLSConfig sets the TLS configuration for the HTTP client.
	SetTLSConfig(tlsConfig *tls.Config) error
}

// HTTPClient is a wrapper around http.Client that implements HTTPClientInterface.
type HTTPClient struct {
	Client *http.Client
}

// Do sends an HTTP request and returns an HTTP response.
func (hc *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	return hc.Client.Do(req)
}

// SetTLSConfig sets the TLS configuration for the HTTP client.
func (hc *HTTPClient) SetTLSConfig(tlsConfig *tls.Config) error {
	if hc.Client == nil {
		return fmt.Errorf("http.Client is nil")
	}
	tr, ok := hc.Client.Transport.(*http.Transport)
	if !ok || tr == nil {
		tr = &http.Transport{}
	}
	tr.TLSClientConfig = tlsConfig
	hc.Client.Transport = tr
	return nil
}

// NotarySigner is responsible for signing images using Notary v2.
type NotarySigner struct {
	URL            string
	RetryTimeout   time.Duration
	PayloadBuilder PayloadBuilderInterface
	TLSProvider    TLSProviderInterface
	HTTPClient     HTTPClientInterface
}

// Sign signs the provided images by sending a signing request to the Notary server.
func (ns *NotarySigner) Sign(images []string) error {
	sImg := strings.Join(images, ", ")

	// Build the signing payload.
	payload, err := ns.PayloadBuilder.BuildPayload(images)
	if err != nil {
		return fmt.Errorf("failed to build payload: %w", err)
	}

	// Marshal the payload into JSON.
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal signing request: %w", err)
	}

	// Obtain TLS configuration.
	tlsConfig, err := ns.TLSProvider.GetTLSConfig()
	if err != nil {
		return fmt.Errorf("failed to get TLS configuration: %w", err)
	}

	// Set TLS configuration for the HTTP client.
	err = ns.HTTPClient.SetTLSConfig(tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to set TLS configuration: %w", err)
	}

	// Create an HTTP POST request with the signing payload.
	req, err := http.NewRequest("POST", ns.URL, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")

	// Send the request with retries.
	resp, err := RetryHTTPRequest(ns.HTTPClient, req, 5, ns.RetryTimeout)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("Successfully signed images: %s\n", sImg)
	return nil
}

// RetryHTTPRequest sends an HTTP request with retry logic in case of failures.
func RetryHTTPRequest(client HTTPClientInterface, req *http.Request, retries int, retryInterval time.Duration) (*http.Response, error) {
	var resp *http.Response
	var err error

	// Read and store the request body for potential retries.
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		req.Body.Close()
	}

	for retries > 0 {
		// Reset the request body for each retry.
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		// Send the HTTP request.
		resp, err = client.Do(req)
		if err != nil {
			// err is already set
		} else if resp.StatusCode == http.StatusAccepted {
			return resp, nil
		} else {
			// Read and discard the response body to free resources
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			err = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			resp = nil // Discard the unsuccessful response
		}

		// Decrement the retry counter.
		retries--
		if retries == 0 {
			break
		}

		// Wait before the next retry.
		time.Sleep(retryInterval)
	}

	return nil, fmt.Errorf("request failed after retries: %w", err)
}

// NewSigner constructs a new NotarySigner with the necessary dependencies.
func (nc *NotaryConfig) NewSigner() (Signer, error) {
	// Read the TLS credentials from the secret file.
	secretFileContent, err := os.ReadFile(nc.Secret.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret file: %w", err)
	}
	var tlsCredentials TLSCredentials
	err = json.Unmarshal(secretFileContent, &tlsCredentials)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal TLS credentials: %w", err)
	}

	// Initialize the TLS provider with the credentials.
	tlsProvider := &TLSProvider{
		Credentials: tlsCredentials,
	}

	// **Validate the TLS credentials**
	_, err = tlsProvider.GetTLSConfig()
	if err != nil {
		return nil, fmt.Errorf("invalid TLS credentials: %w", err)
	}

	// Initialize the image service.
	imageService := &ImageService{}

	// Initialize the payload builder.
	payloadBuilder := &PayloadBuilder{
		ImageService: imageService,
	}

	// Initialize the HTTP client with a timeout.
	httpClient := &HTTPClient{
		Client: &http.Client{
			Timeout: nc.Timeout,
		},
	}

	// Create the NotarySigner with all dependencies injected.
	signer := &NotarySigner{
		URL:            nc.Endpoint,
		RetryTimeout:   nc.RetryTimeout,
		PayloadBuilder: payloadBuilder,
		TLSProvider:    tlsProvider,
		HTTPClient:     httpClient,
	}
	return signer, nil
}

// Target represents an individual image target to be signed.
type Target struct {
	Name     string `json:"name"`
	ByteSize int64  `json:"byteSize"`
	Digest   string `json:"digest"`
}

// GUNTargets associates a GUN with its targets.
type GUNTargets struct {
	GUN     string   `json:"gun"`
	Targets []Target `json:"targets"`
}

// SigningPayload represents the payload to be sent to the Notary server for signing.
type SigningPayload struct {
	GunTargets []GUNTargets `json:"gunTargets"`
}

// TLSCredentials holds the base64-encoded TLS certificate and private key data.
type TLSCredentials struct {
	CertificateData string `json:"certData"`
	PrivateKeyData  string `json:"privateKeyData"`
}

// NotaryConfig holds the configuration for the NotarySigner.
type NotaryConfig struct {
	Endpoint     string            `yaml:"endpoint" json:"endpoint"`
	Secret       *AuthSecretConfig `yaml:"secret,omitempty" json:"secret,omitempty"`
	Timeout      time.Duration     `yaml:"timeout" json:"timeout"`
	RetryTimeout time.Duration     `yaml:"retry-timeout" json:"retry-timeout"`
}

// AuthSecretConfig specifies the path and type of the secret containing TLS credentials.
type AuthSecretConfig struct {
	Path string `yaml:"path" json:"path"`
	Type string `yaml:"type" json:"type"`
}
