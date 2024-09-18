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

// generateTestCert generuje samopodpisany certyfikat i klucz prywatny.
// Zwraca certyfikat i klucz w formacie PEM.
func generateTestCert() (string, string, error) {
	// Generowanie klucza RSA
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	// Tworzenie szablonu certyfikatu
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Organization"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour), // Certyfikat ważny przez 24 godziny
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Samopodpisanie certyfikatu
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return "", "", err
	}

	// Kodowanie certyfikatu do PEM
	certPEM := new(bytes.Buffer)
	if err := pem.Encode(certPEM, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return "", "", err
	}

	// Kodowanie klucza prywatnego do PEM
	keyPEM := new(bytes.Buffer)
	if err := pem.Encode(keyPEM, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}); err != nil {
		return "", "", err
	}

	return certPEM.String(), keyPEM.String(), nil
}

// TestImageService_ParseReference_Valid sprawdza poprawne parsowanie referencji obrazu.
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

// TestImageService_ParseReference_Invalid sprawdza błędne parsowanie referencji obrazu.
func TestImageService_ParseReference_Invalid(t *testing.T) {
	imageService := sign.ImageService{}
	// Użyj bardziej nieprawidłowego formatu referencji
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

// TestImageService_GetImage_Valid sprawdza pobieranie poprawnego obrazu.
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

// TestImageService_GetImage_Invalid sprawdza pobieranie niepoprawnego obrazu.
func TestImageService_GetImage_Invalid(t *testing.T) {
	imageService := sign.ImageService{}
	// Użyj bardziej nieprawidłowego formatu referencji
	ref, err := name.ParseReference("invalid_image")
	if err != nil {
		t.Fatalf("Failed to parse reference: %v", err)
	}
	_, err = imageService.GetImage(ref)
	if err == nil {
		t.Errorf("Expected an error for invalid image")
	}
}

// TestPayloadBuilder_BuildPayload_Valid sprawdza budowanie payloadu dla poprawnych obrazów.
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

// TestPayloadBuilder_BuildPayload_Invalid sprawdza budowanie payloadu dla niepoprawnych obrazów.
func TestPayloadBuilder_BuildPayload_Invalid(t *testing.T) {
	imageService := sign.ImageService{}
	payloadBuilder := sign.PayloadBuilder{ImageService: &imageService}
	_, err := payloadBuilder.BuildPayload([]string{"invalid_image"})
	if err == nil {
		t.Errorf("Expected an error for invalid images")
	}
}

// TestCertificateProvider_CreateKeyPair_Valid sprawdza tworzenie pary kluczy z poprawnymi danymi base64.
func TestCertificateProvider_CreateKeyPair_Valid(t *testing.T) {
	certPEM, keyPEM, err := generateTestCert()
	if err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}

	signifySecret := sign.SignifySecret{
		CertificateData: base64.StdEncoding.EncodeToString([]byte(certPEM)),
		PrivateKeyData:  base64.StdEncoding.EncodeToString([]byte(keyPEM)),
	}
	certificateProvider := sign.CertificateProvider{SignifySecret: signifySecret}
	cert, err := certificateProvider.CreateKeyPair()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	// Sprawdzamy, czy cert nie jest pusty
	if len(cert.Certificate) == 0 {
		t.Errorf("Expected cert.Certificate to have data")
	}
	// Opcjonalnie: Sprawdzenie poprawności certyfikatu
	// Możesz użyć bibliotek takich jak x509 do dalszej weryfikacji
}

// TestCertificateProvider_CreateKeyPair_Invalid sprawdza tworzenie pary kluczy z niepoprawnymi danymi base64.
func TestCertificateProvider_CreateKeyPair_Invalid(t *testing.T) {
	signifySecret := sign.SignifySecret{
		CertificateData: "invalid-base64",
		PrivateKeyData:  base64.StdEncoding.EncodeToString([]byte("private-key-data")),
	}
	certificateProvider := sign.CertificateProvider{SignifySecret: signifySecret}
	_, err := certificateProvider.CreateKeyPair()
	if err == nil {
		t.Errorf("Expected an error for invalid base64 data")
	}
}

// TestTLSConfigurator_SetupTLS sprawdza konfigurację TLS.
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

// TestHTTPClient_Do sprawdza wysyłanie żądania HTTP.
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

// TestHTTPClient_SetTLSConfig sprawdza ustawienie konfiguracji TLS w HTTPClient.
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
	// Sprawdzamy, czy Transport jest ustawiony i jest typu *http.Transport
	transport, ok := httpClient.Client.Transport.(*http.Transport)
	if !ok {
		t.Errorf("Expected Transport to be of type *http.Transport")
	}
	if transport == nil {
		t.Errorf("Expected Transport to be set")
	}
}

// TestNotarySigner_Sign_Valid sprawdza podpisywanie poprawnych obrazów.
func TestNotarySigner_Sign_Valid(t *testing.T) {
	certPEM, keyPEM, err := generateTestCert()
	if err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}

	signifySecret := sign.SignifySecret{
		CertificateData: base64.StdEncoding.EncodeToString([]byte(certPEM)),
		PrivateKeyData:  base64.StdEncoding.EncodeToString([]byte(keyPEM)),
	}
	imageService := sign.ImageService{}
	payloadBuilder := sign.PayloadBuilder{ImageService: &imageService}
	certificateProvider := sign.CertificateProvider{SignifySecret: signifySecret}
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

// TestNotarySigner_Sign_Invalid sprawdza podpisywanie niepoprawnych obrazów.
func TestNotarySigner_Sign_Invalid(t *testing.T) {
	certPEM, keyPEM, err := generateTestCert()
	if err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}

	signifySecret := sign.SignifySecret{
		CertificateData: base64.StdEncoding.EncodeToString([]byte(certPEM)),
		PrivateKeyData:  base64.StdEncoding.EncodeToString([]byte(keyPEM)),
	}
	imageService := sign.ImageService{}
	payloadBuilder := sign.PayloadBuilder{ImageService: &imageService}
	certificateProvider := sign.CertificateProvider{SignifySecret: signifySecret}
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

// TestRetryHTTPRequest_Failure sprawdza, czy RetryHTTPRequest zwraca błąd po nieudanych próbach.
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

// TestRetryHTTPRequest_SuccessAfterRetries sprawdza, czy RetryHTTPRequest kończy sukcesem po określonej liczbie prób.
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
