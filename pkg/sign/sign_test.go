package sign

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestNotaryConfigUnmarshalYAML(t *testing.T) {
	// Mock YAML data with SignerConfig and NotaryConfig
	yamlData := `
name: notary-signer
type: notary
job-type: 
  - postsubmit
config:
  endpoint: https://notary.example.com
  secret:
    path: /path/to/secret
    type: signify
  timeout: 10s
  retry-timeout: 5s
`

	// Unmarshal the YAML into SignerConfig struct
	var sc SignerConfig
	err := yaml.Unmarshal([]byte(yamlData), &sc)
	if err != nil {
		t.Fatalf("failed to unmarshal YAML: %v", err)
	}

	// Assertions for SignerConfig fields
	if sc.Name != "notary-signer" {
		t.Errorf("expected name to be 'notary-signer', got %s", sc.Name)
	}
	if sc.Type != "notary" {
		t.Errorf("expected type to be 'notary', got %s", sc.Type)
	}
	if len(sc.JobType) != 1 || sc.JobType[0] != "postsubmit" {
		t.Errorf("expected job-type to contain 'postsubmit', got %v", sc.JobType)
	}

	// Assertions for NotaryConfig fields (from sc.Config)
	notaryConfig, ok := sc.Config.(*NotaryConfig)
	if !ok {
		t.Fatalf("expected sc.Config to be of type *NotaryConfig, but got %T", sc.Config)
	}

	if notaryConfig.Endpoint != "https://notary.example.com" {
		t.Errorf("expected endpoint to be 'https://notary.example.com', got %s", notaryConfig.Endpoint)
	}
	if notaryConfig.Secret.Path != "/path/to/secret" {
		t.Errorf("expected secret path to be '/path/to/secret', got %s", notaryConfig.Secret.Path)
	}
	if notaryConfig.Secret.Type != "signify" {
		t.Errorf("expected secret type to be 'signify', got %s", notaryConfig.Secret.Type)
	}
	if notaryConfig.Timeout != 10*time.Second {
		t.Errorf("expected timeout to be 10s, got %v", notaryConfig.Timeout)
	}
	if notaryConfig.RetryTimeout != 5*time.Second {
		t.Errorf("expected retry timeout to be 5s, got %v", notaryConfig.RetryTimeout)
	}
}

func TestNotaryConfig_NewSigner(t *testing.T) {
	// Set up a mock secret file in a valid temporary path
	secretPath := "/tmp/mock_secret.json"
	mockSecret := TLSCredentials{
		CertificateData: "mockCertData",
		PrivateKeyData:  "mockPrivateKeyData",
	}
	secretContent, _ := json.Marshal(mockSecret)

	// Write the mock secret to the file
	err := os.WriteFile(secretPath, secretContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write mock secret file: %v", err)
	}
	defer os.Remove(secretPath) // Clean up the file after test

	// Prepare the NotaryConfig with the temporary file path
	config := &NotaryConfig{
		Endpoint:     "https://notary.example.com",
		Secret:       &AuthSecretConfig{Path: secretPath},
		Timeout:      10 * time.Second,
		RetryTimeout: 5 * time.Second,
	}

	signer, err := config.NewSigner()
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}

	if signer == nil {
		t.Errorf("expected a valid signer, but got nil")
	}
}
