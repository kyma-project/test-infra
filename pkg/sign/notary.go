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

// ImageFetcherInterface abstracts fetching images.
type ImageFetcherInterface interface {
	Fetch(ref ReferenceInterface) (ImageInterface, error)
}

// ReferenceParserInterface abstracts parsing image references.
type ReferenceParserInterface interface {
	Parse(image string) (ReferenceInterface, error)
}

// ImageFetcher implements ImageFetcherInterface.
type ImageFetcher struct{}

// ReferenceParser implements ReferenceParserInterface.
type ReferenceParser struct{}

// Fetch fetches the image using the external library.
func (zf *ImageFetcher) Fetch(ref ReferenceInterface) (ImageInterface, error) {
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

func (rp *ReferenceParser) Parse(image string) (ReferenceInterface, error) {
	ref, err := name.ParseReference(image)
	if err != nil {
		return nil, fmt.Errorf("failed to parse image reference: %w", err)
	}
	return &ReferenceWrapper{Ref: ref}, nil
}

// ReferenceInterface abstracts name.Reference
type ReferenceInterface interface {
	Name() string
	String() string
	GetRepositoryName() string
	GetTag() (string, error)
}

type ManifestInterface interface {
	GetConfigSize() int64
	GetConfigDigest() string
}

// ImageInterface abstracts v1.Image
type ImageInterface interface {
	Manifest() (ManifestInterface, error)
}

// ReferenceWrapper wraps name.Reference
type ReferenceWrapper struct {
	Ref name.Reference
}

type ManifestWrapper struct {
	manifest *v1.Manifest
}

func (mw *ManifestWrapper) GetConfigSize() int64 {
	return mw.manifest.Config.Size
}

func (mw *ManifestWrapper) GetConfigDigest() string {
	return mw.manifest.Config.Digest.String()
}

func (rw *ReferenceWrapper) Name() string {
	return rw.Ref.Name()
}

func (rw *ReferenceWrapper) String() string {
	return rw.Ref.String()
}

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

func (rw *ReferenceWrapper) GetTag() (string, error) {
	switch ref := rw.Ref.(type) {
	case name.Tag:
		return ref.TagStr(), nil
	default:
		return "", fmt.Errorf("reference is not a tag")
	}
}

// ImageWrapper wraps v1.Image
type ImageWrapper struct {
	img v1.Image
}

func (iw *ImageWrapper) Manifest() (ManifestInterface, error) {
	manifest, err := iw.img.Manifest()
	if err != nil {
		return nil, err
	}
	return &ManifestWrapper{manifest: manifest}, nil
}

// ImageRepositoryInterface handles image parsing and fetching.
type ImageRepositoryInterface interface {
	ParseReference(image string) (ReferenceInterface, error)
	GetImage(ref ReferenceInterface) (ImageInterface, error)
}

// PayloadBuilderInterface constructs the signing payload.
type PayloadBuilderInterface interface {
	BuildPayload(images []string) (SigningPayload, error)
}

// CertificateProviderInterface manages certificate and key decoding.
type CertificateProviderInterface interface {
	CreateKeyPair() (tls.Certificate, error)
}

// TLSConfiguratorInterface sets up TLS configurations.
type TLSConfiguratorInterface interface {
	SetupTLS(cert tls.Certificate) *tls.Config
}

// HTTPClientInterface handles HTTP requests.
type HTTPClientInterface interface {
	Do(req *http.Request) (*http.Response, error)
	SetTLSConfig(tlsConfig *tls.Config) error
}

type Target struct {
	Name     string `json:"name"`
	ByteSize int64  `json:"byteSize"`
	Digest   string `json:"digest"`
}

type GUNTargets struct {
	GUN     string   `json:"gun"`
	Targets []Target `json:"targets"`
}

type SigningPayload struct {
	GunTargets []GUNTargets `json:"gunTargets"`
}

type TLSCredentials struct {
	CertificateData string `json:"certData"`
	PrivateKeyData  string `json:"privateKeyData"`
}

// NotaryConfig structs
type NotaryConfig struct {
	Endpoint     string            `yaml:"endpoint" json:"endpoint"`
	Secret       *AuthSecretConfig `yaml:"secret,omitempty" json:"secret,omitempty"`
	Timeout      time.Duration     `yaml:"timeout" json:"timeout"`
	RetryTimeout time.Duration     `yaml:"retry-timeout" json:"retry-timeout"`
}

type AuthSecretConfig struct {
	Path string `yaml:"path" json:"path"`
	Type string `yaml:"type" json:"type"`
}

// ImageService implements ImageRepositoryInterface.
type ImageService struct {
	ReferenceParser ReferenceParserInterface
	ImageFetcher    ImageFetcherInterface
}

func (is *ImageService) ParseReference(image string) (ReferenceInterface, error) {
	return is.ReferenceParser.Parse(image)
}

func (is *ImageService) GetImage(ref ReferenceInterface) (ImageInterface, error) {
	img, err := is.ImageFetcher.Fetch(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image: %w", err)
	}
	return img, nil
}

// PayloadBuilder implements PayloadBuilderInterface.
type PayloadBuilder struct {
	ImageService ImageRepositoryInterface
}

func (pb *PayloadBuilder) BuildPayload(images []string) (SigningPayload, error) {
	var gunTargets []GUNTargets
	for _, image := range images {
		// Parse reference
		ref, err := pb.ImageService.ParseReference(image)
		if err != nil {
			return SigningPayload{}, fmt.Errorf("ref parse: %w", err)
		}

		// Get base (repository name) and tag from the reference
		base := ref.GetRepositoryName()
		tag, err := ref.GetTag()
		if err != nil {
			return SigningPayload{}, fmt.Errorf("failed to get tag: %w", err)
		}

		// Get image
		img, err := pb.ImageService.GetImage(ref)
		if err != nil {
			return SigningPayload{}, fmt.Errorf("get image: %w", err)
		}

		// Get manifest
		manifest, err := img.Manifest()
		if err != nil {
			return SigningPayload{}, fmt.Errorf("failed getting image manifest: %w", err)
		}

		// Build target
		target := Target{
			Name:     tag,
			ByteSize: manifest.GetConfigSize(),
			Digest:   manifest.GetConfigDigest(),
		}

		// Build GUN target
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

// CertificateProvider implements CertificateProviderInterface.
type CertificateProvider struct {
	Credentials TLSCredentials
}

func (cp *CertificateProvider) CreateKeyPair() (tls.Certificate, error) {
	certData, err := base64.StdEncoding.DecodeString(cp.Credentials.CertificateData)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to decode certificate: %w", err)
	}
	keyData, err := base64.StdEncoding.DecodeString(cp.Credentials.PrivateKeyData)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to decode private key: %w", err)
	}
	cert, err := tls.X509KeyPair(certData, keyData)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("unable to load certificate or key: %w", err)
	}
	return cert, nil
}

// TLSConfigurator implements TLSConfiguratorInterface.
type TLSConfigurator struct{}

func (tc *TLSConfigurator) SetupTLS(cert tls.Certificate) *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
}

// HTTPClient implements HTTPClientInterface.
type HTTPClient struct {
	Client *http.Client
}

func (hc *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	return hc.Client.Do(req)
}

func (hc *HTTPClient) SetTLSConfig(tlsConfig *tls.Config) error {
	if hc.Client == nil {
		return fmt.Errorf("http.Client is nil")
	}
	hc.Client.Transport = &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	return nil
}

// NotarySigner struct
type NotarySigner struct {
	URL                 string
	RetryTimeout        time.Duration
	PayloadBuilder      PayloadBuilderInterface
	CertificateProvider CertificateProviderInterface
	HTTPClient          HTTPClientInterface
	TLSConfig           *tls.Config
}

// Sign implements the Signer interface.
func (ns *NotarySigner) Sign(images []string) error {
	sImg := strings.Join(images, ", ")

	// Build payload
	payload, err := ns.PayloadBuilder.BuildPayload(images)
	if err != nil {
		return fmt.Errorf("failed to build payload: %v", err)
	}

	// Marshal payload
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal signing request: %v", err)
	}

	// Use the stored TLS configuration
	err = ns.HTTPClient.SetTLSConfig(ns.TLSConfig)
	if err != nil {
		return fmt.Errorf("failed to set TLS config: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", ns.URL, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}
	req.Header.Add("Content-Type", "application/json")

	// Send request with retries
	resp, err := RetryHTTPRequest(ns.HTTPClient, req, 5, ns.RetryTimeout)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		respMsg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to sign images: %s", string(respMsg))
	}

	fmt.Printf("Successfully signed images %s!\n", sImg)
	return nil
}

// RetryHTTPRequest handles retry logic for HTTP requests.
func RetryHTTPRequest(client HTTPClientInterface, req *http.Request, retries int, retryInterval time.Duration) (*http.Response, error) {
	var lastResp *http.Response
	var lastErr error

	// Read and store the request body
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %v", err)
		}
		req.Body.Close()
	}

	for retries > 0 {
		// Reset the request body for each retry
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
		} else if resp.StatusCode == http.StatusAccepted {
			return resp, nil
		} else {
			lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		lastResp = resp
		retries--
		if retries == 0 {
			break
		}
		time.Sleep(retryInterval)
	}
	return lastResp, fmt.Errorf("request failed after retries: %v", lastErr)
}

// NewSigner constructs a new NotarySigner with dependencies injected.
func (nc *NotaryConfig) NewSigner() (Signer, error) {
	// Read secret from the path directly
	secretFileContent, err := os.ReadFile(nc.Secret.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret file: %v", err)
	}
	var tlsCredentials TLSCredentials
	err = json.Unmarshal(secretFileContent, &tlsCredentials)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal TLS credentials: %v", err)
	}

	// Initialize components
	referenceParser := &ReferenceParser{}
	imageFetcher := &ImageFetcher{}
	imageService := &ImageService{
		ReferenceParser: referenceParser,
		ImageFetcher:    imageFetcher,
	}
	payloadBuilder := &PayloadBuilder{
		ImageService: imageService,
	}
	certificateProvider := &CertificateProvider{
		Credentials: tlsCredentials,
	}

	// Create certificate and TLS configuration
	cert, err := certificateProvider.CreateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate and key: %v", err)
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	httpClient := &HTTPClient{
		Client: &http.Client{
			Timeout: nc.Timeout,
		},
	}

	// Create NotarySigner
	signer := &NotarySigner{
		URL:                 nc.Endpoint,
		RetryTimeout:        nc.RetryTimeout,
		PayloadBuilder:      payloadBuilder,
		CertificateProvider: certificateProvider,
		HTTPClient:          httpClient,
		TLSConfig:           tlsConfig,
	}

	return signer, nil
}
