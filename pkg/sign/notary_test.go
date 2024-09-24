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
	"math/big"
	"net/http"
	"net/http/httptest"
	"reflect"
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
	imageService := sign.ImageService{}
	ref, err := imageService.ParseReference("docker.io/library/alpine:latest")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if ref == nil {
		t.Errorf("Expected ref to be not nil")
	}
}

// TestImageService_ParseReference_Invalid checks the incorrect parsing of an image reference.
func TestImageService_ParseReference_Invalid(t *testing.T) {
	imageService := sign.ImageService{}
	// Use more invalid reference formats
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
	imageService := sign.ImageService{}
	ref, err := name.ParseReference("docker.io/library/alpine:latest")
	if err != nil {
		t.Fatalf("Failed to parse reference: %v", err)
	}
	img, err := imageService.GetImage(ref)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if img == nil || reflect.ValueOf(img).IsNil() {
		t.Errorf("Expected img to be not nil")
	}
}

// TestImageService_GetImage_Invalid checks fetching an invalid image.
func TestImageService_GetImage_Invalid(t *testing.T) {
	imageService := sign.ImageService{}
	// Use a more invalid reference format
	ref, err := name.ParseReference("invalid_image")
	if err != nil {
		t.Fatalf("Failed to parse reference: %v", err)
	}
	_, err = imageService.GetImage(ref)
	if err == nil {
		t.Errorf("Expected an error for invalid image")
	}
}

// TestPayloadBuilder_BuildPayload_Valid checks building a payload for valid images.
func TestPayloadBuilder_BuildPayload_Valid(t *testing.T) {
	imageService := sign.ImageService{}
	payloadBuilder := sign.PayloadBuilder{ImageService: &imageService}
	payload, err := payloadBuilder.BuildPayload([]string{"docker.io/library/alpine:latest"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if reflect.DeepEqual(payload, sign.SigningPayload{}) {
		t.Errorf("Expected payload to be not zero value")
	}
}

// TestPayloadBuilder_BuildPayload_Invalid checks building a payload for invalid images.
func TestPayloadBuilder_BuildPayload_Invalid(t *testing.T) {
	imageService := sign.ImageService{}
	payloadBuilder := sign.PayloadBuilder{ImageService: &imageService}
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
	certPEM, keyPEM, err := generateTestCert()
	if err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}

	signifySecret := sign.TLSCredentials{
		CertificateData: base64.StdEncoding.EncodeToString([]byte(certPEM)),
		PrivateKeyData:  base64.StdEncoding.EncodeToString([]byte(keyPEM)),
	}
	imageService := sign.ImageService{}
	payloadBuilder := sign.PayloadBuilder{ImageService: &imageService}
	certificateProvider := sign.CertificateProvider{Credentials: signifySecret}
	tlsConfigurator := sign.TLSConfigurator{}
	httpClient := sign.HTTPClient{Client: &http.Client{}}
	notarySigner := sign.NotarySigner{
		URL:                 "http://example.com",
		RetryTimeout:        1 * time.Second,
		PayloadBuilder:      &payloadBuilder,
		CertificateProvider: &certificateProvider,
		TLSConfigurator:     &tlsConfigurator,
		HTTPClient:          &httpClient,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	notarySigner.URL = server.URL
	err = notarySigner.Sign([]string{"docker.io/library/alpine:latest"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestNotarySigner_Sign_Invalid checks signing invalid images.
func TestNotarySigner_Sign_Invalid(t *testing.T) {
	certPEM, keyPEM, err := generateTestCert()
	if err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}

	signifySecret := sign.TLSCredentials{
		CertificateData: base64.StdEncoding.EncodeToString([]byte(certPEM)),
		PrivateKeyData:  base64.StdEncoding.EncodeToString([]byte(keyPEM)),
	}
	imageService := sign.ImageService{}
	payloadBuilder := sign.PayloadBuilder{ImageService: &imageService}
	certificateProvider := sign.CertificateProvider{Credentials: signifySecret}
	tlsConfigurator := sign.TLSConfigurator{}
	httpClient := sign.HTTPClient{Client: &http.Client{}}
	notarySigner := sign.NotarySigner{
		URL:                 "http://example.com",
		RetryTimeout:        1 * time.Second,
		PayloadBuilder:      &payloadBuilder,
		CertificateProvider: &certificateProvider,
		TLSConfigurator:     &tlsConfigurator,
		HTTPClient:          &httpClient,
	}

	err = notarySigner.Sign([]string{"invalid_image"})
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
	if resp.StatusCode != http.StatusInternalServerError {
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
