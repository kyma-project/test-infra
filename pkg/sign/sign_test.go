package sign

import (
	"fmt"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestNotaryConfigUnmarshalYAML(t *testing.T) {
	yamlContent := `
endpoint: "https://signing.example.com"
secret:
  path: "/path/to/secret"
  type: "signify"
timeout: 30s
retry-timeout: 10s
`
	var notaryConfig NotaryConfig
	err := yaml.Unmarshal([]byte(yamlContent), &notaryConfig)
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}

	if notaryConfig.Endpoint != "https://signing.example.com" {
		t.Errorf("expected endpoint to be %v, but got %v", "https://signing.example.com", notaryConfig.Endpoint)
	}

	if notaryConfig.Secret.Path != "/path/to/secret" {
		t.Errorf("expected secret path to be %v, but got %v", "/path/to/secret", notaryConfig.Secret.Path)
	}

	if notaryConfig.Secret.Type != "signify" {
		t.Errorf("expected secret type to be %v, but got %v", "signify", notaryConfig.Secret.Type)
	}

	if notaryConfig.Timeout != 30*time.Second {
		t.Errorf("expected timeout to be %v, but got %v", 30*time.Second, notaryConfig.Timeout)
	}

	if notaryConfig.RetryTimeout != 10*time.Second {
		t.Errorf("expected retry timeout to be %v, but got %v", 10*time.Second, notaryConfig.RetryTimeout)
	}
}

func TestNotaryConfig_NewSigner(t *testing.T) {
	// Mocking the file reading function to simulate secret loading
	mockReadFileFunc := func(path string) ([]byte, error) {
		if path == "/path/to/secret" {
			return []byte(`{"CertificateData": "mockCertData", "PrivateKeyData": "mockKeyData"}`), nil
		}
		return nil, fmt.Errorf("file not found")
	}

	config := NotaryConfig{
		Endpoint:     "https://signing.example.com",
		Secret:       &AuthSecretConfig{Path: "/path/to/secret", Type: "signify"},
		Timeout:      30 * time.Second,
		RetryTimeout: 10 * time.Second,
		ReadFileFunc: mockReadFileFunc,
	}

	signer, err := config.NewSigner()
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}

	if signer == nil {
		t.Errorf("expected a valid signer, but got nil")
	}
}
