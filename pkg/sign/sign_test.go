package sign

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
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
	certPEM, keyPEM, err := generateTestCert()
	if err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}

	mockSecret := TLSCredentials{
		CertificateData: base64.StdEncoding.EncodeToString([]byte(certPEM)),
		PrivateKeyData:  base64.StdEncoding.EncodeToString([]byte(keyPEM)),
	}

	// Set up a mock secret file in a valid temporary path
	secretPath := "/tmp/mock_secret.json"
	secretContent, _ := json.Marshal(mockSecret)

	// Write the mock secret to the file
	err = os.WriteFile(secretPath, secretContent, 0644)
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

func TestPayloadBuilder_BuildPayload_ManifestList(t *testing.T) {
	ref, err := name.ParseReference("docker.io/library/multiarch:latest")
	if err != nil {
		t.Fatal(err)
	}

	mockManifestList := &MockManifestList{
		MockGetDigest: func() (string, error) { return "manifest-list-digest", nil },
		MockGetSize:   func() (int64, error) { return 4096, nil },
	}

	mockImageRepository := &MockImageRepository{
		MockParseReference: func(image string) (name.Reference, error) {
			return ref, nil
		},
		MockIsManifestList: func(name.Reference) (bool, error) {
			return true, nil
		},
		MockGetManifestList: func(name.Reference) (ManifestListInterface, error) {
			return mockManifestList, nil
		},
	}

	payloadBuilder := PayloadBuilder{
		ImageService: mockImageRepository,
	}

	payload, err := payloadBuilder.BuildPayload([]string{"docker.io/library/multiarch:latest"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedGUN := "index.docker.io/library/multiarch"
	if payload.GunTargets[0].GUN != expectedGUN {
		t.Errorf("Expected GUN '%s', got '%s'", expectedGUN, payload.GunTargets[0].GUN)
	}

	target := payload.GunTargets[0].Targets[0]
	if target.Digest != "manifest-list-digest" {
		t.Errorf("Expected digest 'manifest-list-digest', got '%s'", target.Digest)
	}

	if target.ByteSize != 4096 {
		t.Errorf("Expected size 4096, got %d", target.ByteSize)
	}
}
