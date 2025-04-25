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
	ParseReference(image string) (name.Reference, error)
	// GetImage retrieves an image given its reference.
	GetImage(ref name.Reference) (ImageInterface, error)
	// IsManifestList checks if the reference is a manifest list.
	IsManifestList(ref name.Reference) (bool, error)
	// GetManifestList retrieves the manifest list for the given reference.
	GetManifestList(ref name.Reference) (ManifestListInterface, error)
}

// ManifestListInterface defines methods for working with manifest lists.
type ManifestListInterface interface {
	// GetDigest retrieves the digest of the manifest list.
	GetDigest() (string, error)
	// GetSize retrieves the size of the manifest list.
	GetSize() (int64, error)
}

// ImageInterface abstracts the functionality of v1.Image.
type ImageInterface interface {
	// GetDigest returns the digest of the image manifest.
	GetDigest() (string, error)
	// GetSize returns the size of the image.
	GetSize() (int64, error)
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
func (is *ImageService) ParseReference(image string) (name.Reference, error) {
	ref, err := name.ParseReference(image)
	if err != nil {
		return nil, fmt.Errorf("failed to parse image reference: %w", err)
	}
	return ref, nil
}

// GetImage fetches the image from the remote registry using the provided reference.
func (is *ImageService) GetImage(ref name.Reference) (ImageInterface, error) {
	img, err := remote.Image(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image: %w", err)
	}
	return &ImageWrapper{img: img}, nil
}

// IsManifestList checks if the reference points to a manifest list.
func (is *ImageService) IsManifestList(ref name.Reference) (bool, error) {
	desc, err := remote.Get(ref)
	if err != nil {
		return false, fmt.Errorf("failed to fetch descriptor: %w", err)
	}
	return desc.MediaType.IsIndex(), nil
}

// GetManifestList retrieves the manifest list for the given reference.
func (is *ImageService) GetManifestList(ref name.Reference) (ManifestListInterface, error) {
	isManifestList, err := is.IsManifestList(ref)
	if err != nil {
		return nil, err
	}
	if !isManifestList {
		return nil, fmt.Errorf("reference does not point to a manifest list")
	}

	idx, err := remote.Index(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest list: %w", err)
	}
	return &ManifestListWrapper{idx: idx}, nil
}

// ImageWrapper wraps a v1.Image to implement the ImageInterface.
type ImageWrapper struct {
	img v1.Image
}

// ManifestListWrapper wraps a v1.ImageIndex to implement the ManifestListInterface.
type ManifestListWrapper struct {
	idx v1.ImageIndex
}

// GetDigest returns the digest of the image manifest.
func (iw *ImageWrapper) GetDigest() (string, error) {
	digest, err := iw.img.Digest()
	if err != nil {
		return "", err
	}
	return digest.Hex, nil
}

// GetSize returns the size of the image.
func (iw *ImageWrapper) GetSize() (int64, error) {
	return iw.img.Size()
}

// GetDigest retrieves the digest of the manifest list.
func (mlw *ManifestListWrapper) GetDigest() (string, error) {
	digest, err := mlw.idx.Digest()
	if err != nil {
		return "", fmt.Errorf("failed to get digest: %w", err)
	}
	return digest.Hex, nil
}

// GetSize retrieves the size of the manifest list.
func (mlw *ManifestListWrapper) GetSize() (int64, error) {
	rawManifest, err := mlw.idx.RawManifest()
	if err != nil {
		return 0, fmt.Errorf("failed to get raw manifest: %w", err)
	}
	return int64(len(rawManifest)), nil
}

// ManifestWrapper wraps a v1.Manifest to implement the ManifestInterface.
type ManifestWrapper struct {
	manifest *v1.Manifest
}

// GetConfigSize returns the size of the image config.
func (mw *ManifestWrapper) GetConfigSize() int64 {
	return mw.manifest.Config.Size
}

// GetConfigDigest returns the digest of the image config without the algorithm prefix.
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
		base := ref.Context().Name()
		var tag string
		if tagged, ok := ref.(name.Tag); ok {
			tag = tagged.TagStr()
		} else {
			return SigningPayload{}, fmt.Errorf("reference is not a tag")
		}

		isManifestList, err := pb.ImageService.IsManifestList(ref)
		if err != nil {
			return SigningPayload{}, fmt.Errorf("failed to check if reference is a manifest list: %w", err)
		}

		var digest string
		var size int64

		if isManifestList {
			manifestList, err := pb.ImageService.GetManifestList(ref)
			if err != nil {
				return SigningPayload{}, fmt.Errorf("failed to fetch manifest list: %w", err)
			}

			digest, err = manifestList.GetDigest()
			if err != nil {
				return SigningPayload{}, fmt.Errorf("failed to get manifest list digest: %w", err)
			}

			size, err = manifestList.GetSize()
			if err != nil {
				return SigningPayload{}, fmt.Errorf("failed to get manifest list size: %w", err)
			}
		} else {
			img, err := pb.ImageService.GetImage(ref)
			if err != nil {
				return SigningPayload{}, fmt.Errorf("failed to fetch image: %w", err)
			}

			digest, err = img.GetDigest()
			if err != nil {
				return SigningPayload{}, fmt.Errorf("failed to get image digest: %w", err)
			}

			size, err = img.GetSize()
			if err != nil {
				return SigningPayload{}, fmt.Errorf("failed to get image size: %w", err)
			}
		}

		// Build the target information.
		target := Target{
			Name:     tag,
			ByteSize: size,
			Digest:   digest,
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
	tlsConfig   *tls.Config
}

// GetTLSConfig constructs a tls.Config using the stored TLS credentials.
func (tp *TLSProvider) GetTLSConfig() (*tls.Config, error) {
	if tp.tlsConfig != nil {
		return tp.tlsConfig, nil
	}
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
	tp.tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	return tp.tlsConfig, nil
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

// NotarySigner is responsible for signing images
type NotarySigner struct {
	url            string
	retryTimeout   time.Duration
	payloadBuilder PayloadBuilderInterface
	tlsProvider    TLSProviderInterface
	httpClient     HTTPClientInterface
}

// Sign signs the provided images by sending a signing request to the Notary server.
func (ns *NotarySigner) Sign(images []string) error {
	sImg := strings.Join(images, ", ")

	// Build the signing payload.
	payload, err := ns.payloadBuilder.BuildPayload(images)
	if err != nil {
		return fmt.Errorf("failed to build payload: %w", err)
	}

	// Marshal the payload into JSON.
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal signing request: %w", err)
	}

	// Obtain TLS configuration.
	tlsConfig, err := ns.tlsProvider.GetTLSConfig()
	if err != nil {
		return fmt.Errorf("failed to get TLS configuration: %w", err)
	}

	// Set TLS configuration for the HTTP client.
	err = ns.httpClient.SetTLSConfig(tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to set TLS configuration: %w", err)
	}

	// Create an HTTP POST request with the signing payload.
	req, err := http.NewRequest("POST", ns.url, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")

	// Send the request with retries.
	resp, err := RetryHTTPRequest(ns.httpClient, req, 5, ns.retryTimeout)
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
			// Continue to the next retry attempt
			time.Sleep(retryInterval)
			retries--
			continue
		}

		if resp.StatusCode == http.StatusAccepted {
			return resp, nil
		}

		// Read and discard the response body to free resources
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		err = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		resp = nil // Discard the unsuccessful response

		// Decrement the retry counter and wait before the next retry.
		retries--
		if retries > 0 {
			time.Sleep(retryInterval)
		}
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

	// Load the TLS credentials
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
		url:            nc.Endpoint,
		retryTimeout:   nc.RetryTimeout,
		payloadBuilder: payloadBuilder,
		tlsProvider:    tlsProvider,
		httpClient:     httpClient,
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
	GunTargets []GUNTargets `json:"trustedCollections"`
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
