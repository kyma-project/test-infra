package sign

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
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
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(24 * time.Hour), // Certificate valid for 24 hours
		KeyUsage:  x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
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
	imageService := ImageService{}

	// Use a valid image reference
	ref, err := imageService.ParseReference("docker.io/library/alpine:latest")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if ref == nil {
		t.Errorf("Expected ref to be non-nil")
	}
}

// TestImageService_ParseReference_Invalid checks the incorrect parsing of an image reference.
func TestImageService_ParseReference_Invalid(t *testing.T) {
	imageService := ImageService{}

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
	ref, err := name.ParseReference("docker.io/library/alpine:latest")
	if err != nil {
		t.Fatal(err)
	}

	mockImageRepository := &MockImageRepository{
		MockParseReference: func(image string) (name.Reference, error) {
			return ref, nil
		},
		MockGetImage: func(ref name.Reference) (ImageInterface, error) {
			return &MockImage{
				MockGetDigest: func() (string, error) { return "dummy-digest", nil },
				MockGetSize:   func() (int64, error) { return 1024, nil },
			}, nil
		},
	}

	// Use mockImageRepository as imageService
	imageService := mockImageRepository

	img, err := imageService.GetImage(ref)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Additional checks for new methods
	digest, err := img.GetDigest()
	if err != nil {
		t.Errorf("Expected no error getting digest, got %v", err)
	}
	if digest != "dummy-digest" {
		t.Errorf("Expected digest to be 'dummy-digest', got '%s'", digest)
	}

	size, err := img.GetSize()
	if err != nil {
		t.Errorf("Expected no error getting size, got %v", err)
	}
	if size != 1024 {
		t.Errorf("Expected size to be 1024, got %d", size)
	}
}

// TestImageService_GetImage_Invalid checks fetching an invalid image.
func TestImageService_GetImage_Invalid(t *testing.T) {
	ref, err := name.ParseReference("invalid_image:latest")
	if err != nil {
		t.Fatal(err)
	}

	mockImageRepository := &MockImageRepository{
		MockParseReference: func(image string) (name.Reference, error) {
			return ref, nil
		},
		MockGetImage: func(ref name.Reference) (ImageInterface, error) {
			return nil, fmt.Errorf("failed to fetch image")
		},
	}

	imageService := mockImageRepository

	_, err = imageService.GetImage(ref)
	if err == nil {
		t.Errorf("Expected an error for invalid image")
	}
}

// TestPayloadBuilder_BuildPayload_Valid checks building a payload for valid images.
func TestPayloadBuilder_BuildPayload_Valid(t *testing.T) {
	t.Run("single image", func(t *testing.T) {
		ref, err := name.ParseReference("docker.io/library/alpine:latest")
		if err != nil {
			t.Fatal(err)
		}

		mockImageRepository := &MockImageRepository{
			MockParseReference: func(image string) (name.Reference, error) {
				return ref, nil
			},
			MockGetImage: func(ref name.Reference) (ImageInterface, error) {
				return &MockImage{
					MockGetDigest: func() (string, error) { return "dummy-manifest-digest", nil },
					MockGetSize:   func() (int64, error) { return 2048, nil },
				}, nil
			},
			MockIsManifestList: func(name.Reference) (bool, error) { return false, nil },
		}

		payloadBuilder := PayloadBuilder{
			ImageService: mockImageRepository,
		}

		payload, err := payloadBuilder.BuildPayload([]string{"docker.io/library/alpine:latest"})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		expectedGUN := "index.docker.io/library/alpine"
		if payload.GunTargets[0].GUN != expectedGUN {
			t.Errorf("Expected GUN '%s', got '%s'", expectedGUN, payload.GunTargets[0].GUN)
		}
	})
}

// TestPayloadBuilder_BuildPayload_Invalid checks building a payload for invalid images.
func TestPayloadBuilder_BuildPayload_Invalid(t *testing.T) {
	mockImageRepository := &MockImageRepository{
		MockParseReference: func(image string) (name.Reference, error) {
			return nil, fmt.Errorf("failed to parse reference")
		},
	}

	payloadBuilder := PayloadBuilder{
		ImageService: mockImageRepository,
	}

	_, err := payloadBuilder.BuildPayload([]string{"invalid_image"})
	if err == nil {
		t.Errorf("Expected an error for invalid images")
	}
}

// TestTLSProvider_GetTLSConfig_Valid checks getting TLS configuration with valid credentials.
func TestTLSProvider_GetTLSConfig_Valid(t *testing.T) {
	certPEM, keyPEM, err := generateTestCert()
	if err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}

	tlsCredentials := TLSCredentials{
		CertificateData: base64.StdEncoding.EncodeToString([]byte(certPEM)),
		PrivateKeyData:  base64.StdEncoding.EncodeToString([]byte(keyPEM)),
	}

	tlsProvider := TLSProvider{Credentials: tlsCredentials}
	tlsConfig, err := tlsProvider.GetTLSConfig()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if tlsConfig == nil {
		t.Errorf("Expected tlsConfig to be not nil")
	}
	if len(tlsConfig.Certificates) != 1 {
		t.Errorf("Expected Certificates length to be 1, got %d", len(tlsConfig.Certificates))
	}
}

// TestTLSProvider_GetTLSConfig_Invalid checks getting TLS configuration with invalid credentials.
func TestTLSProvider_GetTLSConfig_Invalid(t *testing.T) {
	tlsCredentials := TLSCredentials{
		CertificateData: "invalid-base64",
		PrivateKeyData:  base64.StdEncoding.EncodeToString([]byte("private-key-data")),
	}

	tlsProvider := TLSProvider{Credentials: tlsCredentials}
	_, err := tlsProvider.GetTLSConfig()
	if err == nil {
		t.Errorf("Expected an error for invalid credentials")
	}
}

// TestHTTPClient_Do checks sending an HTTP request.
func TestHTTPClient_Do(t *testing.T) {
	httpClient := HTTPClient{Client: &http.Client{}}
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
	httpClient := HTTPClient{Client: &http.Client{}}
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
	mockPayloadBuilder := &MockPayloadBuilder{
		MockBuildPayload: func(images []string) (SigningPayload, error) {
			return SigningPayload{
				GunTargets: []GUNTargets{
					{
						GUN: "docker.io/library/alpine",
						Targets: []Target{
							{
								Name:     "latest",
								ByteSize: 1024,
								Digest:   "dummy-manifest-digest",
							},
						},
					},
				},
			}, nil
		},
	}

	mockTLSProvider := &MockTLSProvider{
		MockGetTLSConfig: func() (*tls.Config, error) {
			certPEM, keyPEM, _ := generateTestCert()
			cert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
			if err != nil {
				return nil, err
			}
			return &tls.Config{Certificates: []tls.Certificate{cert}}, nil
		},
	}

	mockHTTPClient := &MockHTTPClient{
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

	notarySigner := NotarySigner{
		url:            "http://example.com",
		retryTimeout:   1 * time.Second,
		payloadBuilder: mockPayloadBuilder,
		tlsProvider:    mockTLSProvider,
		httpClient:     mockHTTPClient,
	}

	err := notarySigner.Sign([]string{"docker.io/library/alpine:latest"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestNotarySigner_Sign_Invalid checks signing invalid images.
func TestNotarySigner_Sign_Invalid(t *testing.T) {
	mockPayloadBuilder := &MockPayloadBuilder{
		MockBuildPayload: func(images []string) (SigningPayload, error) {
			return SigningPayload{}, fmt.Errorf("failed to build payload")
		},
	}

	mockTLSProvider := &MockTLSProvider{
		MockGetTLSConfig: func() (*tls.Config, error) {
			certPEM, keyPEM, _ := generateTestCert()
			cert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
			if err != nil {
				return nil, err
			}
			return &tls.Config{Certificates: []tls.Certificate{cert}}, nil
		},
	}

	mockHTTPClient := &MockHTTPClient{
		MockDo: func(req *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("request failed")
		},
		MockSetTLSConfig: func(tlsConfig *tls.Config) error {
			return nil
		},
	}

	notarySigner := NotarySigner{
		url:            "http://example.com",
		retryTimeout:   1 * time.Second,
		payloadBuilder: mockPayloadBuilder,
		tlsProvider:    mockTLSProvider,
		httpClient:     mockHTTPClient,
	}

	err := notarySigner.Sign([]string{"invalid_image"})
	if err == nil {
		t.Errorf("Expected an error for invalid images")
	}
}

// TestRetryHTTPRequest_Failure checks if RetryHTTPRequest returns an error after failed attempts.
func TestRetryHTTPRequest_Failure(t *testing.T) {
	httpClient := HTTPClient{Client: &http.Client{}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	resp, err := RetryHTTPRequest(&httpClient, req, 3, 1*time.Second)
	if err == nil {
		t.Errorf("Expected an error on failure")
	}
	if resp != nil && resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
}

// TestRetryHTTPRequest_SuccessAfterRetries checks if RetryHTTPRequest succeeds after a certain number of retries.
func TestRetryHTTPRequest_SuccessAfterRetries(t *testing.T) {
	httpClient := HTTPClient{Client: &http.Client{}}
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
	resp, err := RetryHTTPRequest(&httpClient, req, 3, 1*time.Second)
	if err != nil {
		t.Errorf("Expected no error after retries, got %v", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("Expected status %d, got %d", http.StatusAccepted, resp.StatusCode)
	}
}

func TestNotaryConfig_NewSigner_Success(t *testing.T) {
	// Prepare a temporary file with valid TLS data
	certPEM, keyPEM, err := generateTestCert()
	if err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}

	tlsCredentials := TLSCredentials{
		CertificateData: base64.StdEncoding.EncodeToString([]byte(certPEM)),
		PrivateKeyData:  base64.StdEncoding.EncodeToString([]byte(keyPEM)),
	}

	// Serialize TLS data to JSON format
	secretContent, err := json.Marshal(tlsCredentials)
	if err != nil {
		t.Fatalf("Failed to marshal TLS credentials: %v", err)
	}

	// Create a temporary file for the secret
	tempFile, err := ioutil.TempFile("", "secret-*.json")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write TLS data to the file
	if _, err := tempFile.Write(secretContent); err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}
	tempFile.Close() // Close the file to ensure the data is written

	// Create NotaryConfig with the path to the secret file
	notaryConfig := &NotaryConfig{
		Endpoint:     "http://example.com",
		Secret:       &AuthSecretConfig{Path: tempFile.Name()},
		Timeout:      10 * time.Second,
		RetryTimeout: 1 * time.Second,
	}

	// Call the NewSigner method
	signer, err := notaryConfig.NewSigner()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if signer == nil {
		t.Errorf("Expected signer to be not nil")
	}

	// Optionally, check if signer is of type *NotarySigner
	notarySigner, ok := signer.(*NotarySigner)
	if !ok {
		t.Errorf("Expected signer to be of type *NotarySigner")
	}

	// Additional checks for NotarySigner fields
	if notarySigner.url != notaryConfig.Endpoint {
		t.Errorf("Expected NotarySigner URL to be '%s', got '%s'", notaryConfig.Endpoint, notarySigner.url)
	}
}

func TestNotaryConfig_NewSigner_InvalidSecretFile(t *testing.T) {
	// Create NotaryConfig with a non-existent secret file
	notaryConfig := &NotaryConfig{
		Endpoint:     "http://example.com",
		Secret:       &AuthSecretConfig{Path: "non-existent-file.json"},
		Timeout:      10 * time.Second,
		RetryTimeout: 1 * time.Second,
	}

	// Call the NewSigner method
	signer, err := notaryConfig.NewSigner()
	if err == nil {
		t.Errorf("Expected error due to invalid secret file, got nil")
	}
	if signer != nil {
		t.Errorf("Expected signer to be nil due to error")
	}
}

func TestNotaryConfig_NewSigner_InvalidTLSCredentials(t *testing.T) {
	// Prepare invalid TLS data
	secretContent := []byte(`{"certData": "invalid", "privateKeyData": "also-invalid"}`)

	// Create a temporary file for the secret
	tempFile, err := ioutil.TempFile("", "secret-invalid-*.json")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write invalid TLS data to the file
	if _, err := tempFile.Write(secretContent); err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}
	tempFile.Close()

	// Create NotaryConfig with the path to the secret file
	notaryConfig := &NotaryConfig{
		Endpoint:     "http://example.com",
		Secret:       &AuthSecretConfig{Path: tempFile.Name()},
		Timeout:      10 * time.Second,
		RetryTimeout: 1 * time.Second,
	}

	// Call the NewSigner method
	signer, err := notaryConfig.NewSigner()
	if err == nil {
		t.Errorf("Expected error due to invalid TLS credentials, got nil")
	}
	if signer != nil {
		t.Errorf("Expected signer to be nil due to error")
	}
}

func TestImageService_IsManifestList(t *testing.T) {
	mockRepo := &MockImageRepository{
		MockIsManifestList: func(ref name.Reference) (bool, error) {
			return true, nil
		},
	}

	ref, err := name.ParseReference("docker.io/library/alpine:latest")
	if err != nil {
		t.Fatal(err)
	}

	isList, err := mockRepo.IsManifestList(ref)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !isList {
		t.Error("Expected manifest list detection")
	}
}

func TestNotarySigner_Sign_ManifestList(t *testing.T) {
	mockHTTPClient := &MockHTTPClient{
		MockDo: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusAccepted,
				Body:       io.NopCloser(bytes.NewReader([]byte{})),
			}, nil
		},
		MockSetTLSConfig: func(tlsConfig *tls.Config) error {
			return nil
		},
	}

	mockPayloadBuilder := &MockPayloadBuilder{
		MockBuildPayload: func(images []string) (SigningPayload, error) {
			return SigningPayload{
				GunTargets: []GUNTargets{{
					GUN: "docker.io/library/multiarch",
					Targets: []Target{{
						Name:     "latest",
						ByteSize: 4096,
						Digest:   "manifest-list-digest",
					}},
				}},
			}, nil
		},
	}

	mockTLSProvider := &MockTLSProvider{
		MockGetTLSConfig: func() (*tls.Config, error) {
			return &tls.Config{}, nil
		},
	}

	notarySigner := NotarySigner{
		payloadBuilder: mockPayloadBuilder,
		tlsProvider:    mockTLSProvider,
		httpClient:     mockHTTPClient,
		url:            "http://example.com",
		retryTimeout:   1 * time.Second,
	}

	err := notarySigner.Sign([]string{"docker.io/library/multiarch:latest"})
	if err != nil {
		t.Fatalf("Signing failed: %v", err)
	}
}
