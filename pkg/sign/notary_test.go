package sign_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/kyma-project/test-infra/pkg/sign"
)

// generateTestCert generates a self-signed certificate and private key.
// Returns the certificate and key in PEM format.
func generateTestCert() (string, string, error) {
	// Generate RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Organization"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour), // Certificate valid for 24 hours
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Self-sign the certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return "", "", err
	}

	// Encode certificate to PEM
	certPEM := new(bytes.Buffer)
	if err := pem.Encode(certPEM, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return "", "", err
	}

	// Encode private key to PEM
	keyPEM := new(bytes.Buffer)
	if err := pem.Encode(keyPEM, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}); err != nil {
		return "", "", err
	}

	return certPEM.String(), keyPEM.String(), nil
}

// TestImageService_ParseReference_Valid checks the correct parsing of an image reference.
func TestImageService_ParseReference_Valid(t *testing.T) {
	mockRef := &sign.ReferenceWrapper{} // Możesz dostosować według potrzeb

	mockReferenceParser := &sign.MockReferenceParser{
		MockParse: func(image string) (sign.ReferenceInterface, error) {
			if image == "docker.io/library/alpine:latest" {
				return mockRef, nil
			}
			return nil, fmt.Errorf("invalid image")
		},
	}

	imageService := sign.ImageService{
		ReferenceParser: mockReferenceParser,
	}

	ref, err := imageService.ParseReference("docker.io/library/alpine:latest")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if ref != mockRef {
		t.Errorf("Expected ref to be mockRef")
	}
}

// TestImageService_ParseReference_Invalid checks the incorrect parsing of an image reference.
func TestImageService_ParseReference_Invalid(t *testing.T) {
	mockReferenceParser := &sign.MockReferenceParser{
		MockParse: func(image string) (sign.ReferenceInterface, error) {
			return nil, fmt.Errorf("invalid image reference")
		},
	}

	imageService := sign.ImageService{
		ReferenceParser: mockReferenceParser,
	}

	invalidReferences := []string{
		":::",
		"invalid_image@sha256:invaliddigest",
		"invalid image",
		"@invalid",
		"invalid_image@sha256:",
	}
	for _, image := range invalidReferences {
		_, err := imageService.ParseReference(image)
		if err == nil {
			t.Errorf("Expected an error for invalid image reference '%s', but got nil", image)
		}
	}
}

// TestImageService_GetImage_Valid checks fetching a valid image.
func TestImageService_GetImage_Valid(t *testing.T) {
	mockRef := &sign.ReferenceWrapper{}
	mockManifest := &sign.MockManifest{
		MockGetConfigSize: func() int64 {
			return 1024
		},
		MockGetConfigDigest: func() string {
			return "sha256:dummy-digest"
		},
	}
	mockImage := &sign.MockImage{
		MockManifest: func() (sign.ManifestInterface, error) {
			return mockManifest, nil
		},
	}
	mockImageFetcher := &sign.MockImageFetcher{
		MockFetch: func(ref sign.ReferenceInterface) (sign.ImageInterface, error) {
			if ref == mockRef {
				return mockImage, nil
			}
			return nil, fmt.Errorf("image not found")
		},
	}

	imageService := sign.ImageService{
		ImageFetcher: mockImageFetcher,
	}

	img, err := imageService.GetImage(mockRef)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if img != mockImage {
		t.Errorf("Expected img to be mockImage")
	}
	manifest, err := img.Manifest()
	if err != nil {
		t.Errorf("Expected no error getting manifest, got %v", err)
	}
	if manifest != mockManifest {
		t.Errorf("Expected manifest to be mockManifest")
	}
}

// TestImageService_GetImage_Invalid checks fetching an invalid image.
func TestImageService_GetImage_Invalid(t *testing.T) {
	mockRef := &sign.ReferenceWrapper{}

	mockImageFetcher := &sign.MockImageFetcher{
		MockFetch: func(ref sign.ReferenceInterface) (sign.ImageInterface, error) {
			return nil, fmt.Errorf("failed to fetch image")
		},
	}

	imageService := sign.ImageService{
		ImageFetcher: mockImageFetcher,
	}

	_, err := imageService.GetImage(mockRef)
	if err == nil {
		t.Errorf("Expected an error for invalid image")
	}
}

// TestPayloadBuilder_BuildPayload_Valid checks building a payload for valid images.
func TestPayloadBuilder_BuildPayload_Valid(t *testing.T) {
	tag, err := name.NewTag("docker.io/library/alpine:latest")
	if err != nil {
		t.Fatalf("Failed to create name.Tag: %v", err)
	}
	mockRef := &sign.ReferenceWrapper{Ref: tag}
	mockManifest := &sign.MockManifest{
		MockGetConfigSize: func() int64 {
			return 1024
		},
		MockGetConfigDigest: func() string {
			return "sha256:dummy-digest"
		},
	}
	mockImage := &sign.MockImage{
		MockManifest: func() (sign.ManifestInterface, error) {
			return mockManifest, nil
		},
	}
	mockImageService := &sign.MockImageRepository{
		MockParseReference: func(image string) (sign.ReferenceInterface, error) {
			return mockRef, nil
		},
		MockGetImage: func(ref sign.ReferenceInterface) (sign.ImageInterface, error) {
			return mockImage, nil
		},
	}

	payloadBuilder := sign.PayloadBuilder{
		ImageService: mockImageService,
	}

	payload, err := payloadBuilder.BuildPayload([]string{"docker.io/library/alpine:latest"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(payload.GunTargets) == 0 {
		t.Errorf("Expected GunTargets to be populated")
	}
}

// TestPayloadBuilder_BuildPayload_Invalid checks building a payload for invalid images.
func TestPayloadBuilder_BuildPayload_Invalid(t *testing.T) {
	mockImageService := &sign.MockImageRepository{
		MockParseReference: func(image string) (sign.ReferenceInterface, error) {
			return nil, fmt.Errorf("failed to parse reference")
		},
	}

	payloadBuilder := sign.PayloadBuilder{
		ImageService: mockImageService,
	}

	_, err := payloadBuilder.BuildPayload([]string{"invalid_image"})
	if err == nil {
		t.Errorf("Expected an error for invalid images")
	}
}

// TestCertificateProvider_CreateKeyPair_Valid checks creating a key pair with valid base64 data.
func TestCertificateProvider_CreateKeyPair_Valid(t *testing.T) {
	certPEM, keyPEM, err := generateTestCert()
	if err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}

	signifySecret := sign.TLSCredentials{
		CertificateData: base64.StdEncoding.EncodeToString([]byte(certPEM)),
		PrivateKeyData:  base64.StdEncoding.EncodeToString([]byte(keyPEM)),
	}
	certificateProvider := sign.CertificateProvider{Credentials: signifySecret}
	cert, err := certificateProvider.CreateKeyPair()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	// Check if cert is not empty
	if len(cert.Certificate) == 0 {
		t.Errorf("Expected cert.Certificate to have data")
	}
	// Optionally: Check the correctness of the certificate
	// You can use libraries like x509 for further verification
}

// TestCertificateProvider_CreateKeyPair_Invalid checks creating a key pair with invalid base64 data.
func TestCertificateProvider_CreateKeyPair_Invalid(t *testing.T) {
	signifySecret := sign.TLSCredentials{
		CertificateData: "invalid-base64",
		PrivateKeyData:  base64.StdEncoding.EncodeToString([]byte("private-key-data")),
	}
	certificateProvider := sign.CertificateProvider{Credentials: signifySecret}
	_, err := certificateProvider.CreateKeyPair()
	if err == nil {
		t.Errorf("Expected an error for invalid base64 data")
	}
}

// TestTLSConfigurator_SetupTLS checks TLS configuration.
func TestTLSConfigurator_SetupTLS(t *testing.T) {
	certPEM, keyPEM, err := generateTestCert()
	if err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}

	certificate, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
	if err != nil {
		t.Fatalf("Failed to load X509 key pair: %v", err)
	}

	tlsConfigurator := sign.TLSConfigurator{}
	tlsConfig := tlsConfigurator.SetupTLS(certificate)
	if tlsConfig == nil {
		t.Errorf("Expected tlsConfig to be not nil")
	}
	if len(tlsConfig.Certificates) != 1 {
		t.Errorf("Expected Certificates length to be 1, got %d", len(tlsConfig.Certificates))
	}
}

// TestHTTPClient_Do checks sending an HTTP request.
func TestHTTPClient_Do(t *testing.T) {
	httpClient := sign.HTTPClient{Client: &http.Client{}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

// TestHTTPClient_SetTLSConfig checks setting TLS configuration in HTTPClient.
func TestHTTPClient_SetTLSConfig(t *testing.T) {
	certPEM, keyPEM, err := generateTestCert()
	if err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}

	certificate, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
	if err != nil {
		t.Fatalf("Failed to load X509 key pair: %v", err)
	}

	httpClient := sign.HTTPClient{Client: &http.Client{}}
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{certificate}}
	err = httpClient.SetTLSConfig(tlsConfig)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	// Check if Transport is set and is of type *http.Transport
	transport, ok := httpClient.Client.Transport.(*http.Transport)
	if !ok {
		t.Errorf("Expected Transport to be of type *http.Transport")
	}
	if transport == nil {
		t.Errorf("Expected Transport to be set")
	}
}

// TestNotarySigner_Sign_Valid checks signing valid images.
func TestNotarySigner_Sign_Valid(t *testing.T) {
	mockPayloadBuilder := &sign.MockPayloadBuilder{
		MockBuildPayload: func(images []string) (sign.SigningPayload, error) {
			return sign.SigningPayload{
				GunTargets: []sign.GUNTargets{
					{
						GUN: "docker.io/library/alpine",
						Targets: []sign.Target{
							{
								Name:     "latest",
								ByteSize: 1024,
								Digest:   "sha256:dummy-digest",
							},
						},
					},
				},
			}, nil
		},
	}

	mockCertificateProvider := &sign.MockCertificateProvider{
		MockCreateKeyPair: func() (tls.Certificate, error) {
			certPEM, keyPEM, _ := generateTestCert()
			return tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
		},
	}

	mockHTTPClient := &sign.MockHTTPClient{
		MockDo: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusAccepted,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte{})),
			}, nil
		},
		MockSetTLSConfig: func(tlsConfig *tls.Config) error {
			return nil
		},
	}

	notarySigner := sign.NotarySigner{
		URL:                 "http://example.com",
		RetryTimeout:        1 * time.Second,
		PayloadBuilder:      mockPayloadBuilder,
		CertificateProvider: mockCertificateProvider,
		HTTPClient:          mockHTTPClient,
		TLSConfig:           &tls.Config{},
	}

	err := notarySigner.Sign([]string{"docker.io/library/alpine:latest"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestNotarySigner_Sign_Invalid checks signing invalid images.
func TestNotarySigner_Sign_Invalid(t *testing.T) {
	mockPayloadBuilder := &sign.MockPayloadBuilder{
		MockBuildPayload: func(images []string) (sign.SigningPayload, error) {
			return sign.SigningPayload{}, fmt.Errorf("failed to build payload")
		},
	}

	mockCertificateProvider := &sign.MockCertificateProvider{
		MockCreateKeyPair: func() (tls.Certificate, error) {
			certPEM, keyPEM, _ := generateTestCert()
			return tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
		},
	}

	mockHTTPClient := &sign.MockHTTPClient{
		MockDo: func(req *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("request failed")
		},
		MockSetTLSConfig: func(tlsConfig *tls.Config) error {
			return nil
		},
	}

	notarySigner := sign.NotarySigner{
		URL:                 "http://example.com",
		RetryTimeout:        1 * time.Second,
		PayloadBuilder:      mockPayloadBuilder,
		CertificateProvider: mockCertificateProvider,
		HTTPClient:          mockHTTPClient,
		TLSConfig:           &tls.Config{},
	}

	err := notarySigner.Sign([]string{"invalid_image"})
	if err == nil {
		t.Errorf("Expected an error for invalid images")
	}
}

// TestRetryHTTPRequest_Failure checks if RetryHTTPRequest returns an error after failed attempts.
func TestRetryHTTPRequest_Failure(t *testing.T) {
	httpClient := sign.HTTPClient{Client: &http.Client{}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	resp, err := sign.RetryHTTPRequest(&httpClient, req, 3, 1*time.Second)
	if err == nil {
		t.Errorf("Expected an error on failure")
	}
	if resp != nil && resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
}

// TestRetryHTTPRequest_SuccessAfterRetries checks if RetryHTTPRequest succeeds after a certain number of retries.
func TestRetryHTTPRequest_SuccessAfterRetries(t *testing.T) {
	httpClient := sign.HTTPClient{Client: &http.Client{}}
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusAccepted)
		}
	}))
	defer server.Close()

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	resp, err := sign.RetryHTTPRequest(&httpClient, req, 3, 1*time.Second)
	if err != nil {
		t.Errorf("Expected no error after retries, got %v", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("Expected status %d, got %d", http.StatusAccepted, resp.StatusCode)
	}
}
