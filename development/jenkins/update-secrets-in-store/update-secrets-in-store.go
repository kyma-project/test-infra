package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/kyma-project/test-infra/development/gcp/pkg/secretmanager"
	"google.golang.org/api/option"
	gcpsecretmanager "google.golang.org/api/secretmanager/v1"
	"gopkg.in/yaml.v2"
)

/*
secret file

<org.jenkinsci.plugins.plaincredentials.impl.FileCredentialsImpl plugin="plain-credentials@139.ved2b_9cf7587b">
	<id>secret-manager-kymasecurity-jaas_sap-kyma-prow</id>
	<description></description>
	<fileName>sap-kyma-prow-03cf1cccd8eb.json</fileName>
	<secretBytes>
		<secret-redacted/>
	</secretBytes>
</org.jenkinsci.plugins.plaincredentials.impl.FileCredentialsImpl>

secret text

<org.jenkinsci.plugins.plaincredentials.impl.StringCredentialsImpl plugin="plain-credentials@139.ved2b_9cf7587b">
<id>test-secret-service-account-sap-kyma-prow</id>
<description/>
<secret>
<secret-redacted/>
</secret>
</org.jenkinsci.plugins.plaincredentials.impl.StringCredentialsImpl>
*/

var (
	secretManagerService *secretmanager.Service
	err                  error
	httpClient           *http.Client
)

type CredentialInterface interface {
	UpdateSecret(string) error
}
type SecretTextCredential struct {
	XMLName xml.Name `xml:"org.jenkinsci.plugins.plaincredentials.impl.StringCredentialsImpl"`
	Secret  string   `xml:"secret"`
	Credential
}

func (stc *SecretTextCredential) UpdateSecret(secret string) error {
	stc.Secret = secret
	return nil
}

type SecretFileCredential struct {
	XMLName     xml.Name `xml:"org.jenkinsci.plugins.plaincredentials.impl.FileCredentialsImpl"`
	SecretBytes string   `xml:"secretBytes"`
	Credential
}

func (sfc *SecretFileCredential) UpdateSecret(secret string) error {
	sfc.SecretBytes = secret
	return nil
}

type Credential struct {
	Plugin      string `xml:"plugin,attr"`
	ID          string `xml:"id"`
	Description string `xml:"description"`
	FileName    string `xml:"fileName"`
	// SecretBytes string `xml:"secretBytes"`
	// Secret string `xml:"secret"`
}

func (c *Credential) UpdateSecret(_ string) error {
	return fmt.Errorf("secret update failed, unknown secret type")
}

// type SecretBytes struct {
//	XMLName        xml.Name `xml:"secretBytes"`
//	SecretRedacted string   `xml:"secret-redacted"`
// }

type SyncSecretEvent struct {
	SecretName      string   `yaml:"secret.name,omitempty"`
	SecretVersion   int      `yaml:"secret.version,omitempty"`
	SecretEndpoints []string `yaml:"secret.endpoints,omitempty"`
}

func main() {
	ctx := context.Background()
	secretManagerCreds := os.Getenv("SECRET_MANAGER_JSON")
	changedFiles := os.Getenv("CHANGED_FILES")
	apiToken := os.Getenv("API_TOKEN")
	apiUser := os.Getenv("API_USER")
	// TODO: This is to test jenkins credential value was updated. Remove after POC phase.
	testVar := os.Getenv("TEST_VAR")
	credsOption := option.WithCredentialsJSON([]byte(secretManagerCreds))
	secretManagerService, err = secretmanager.NewService(ctx, credsOption)
	if err != nil {
		panic("failed creating Secret Manager client, error: " + err.Error())
	}
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	httpClient = &http.Client{Timeout: 15 * time.Second, Transport: customTransport}
	// TODO: This is to test jenkins credential value was updated. Remove after POC phase.
	if testVar != "" {
		fmt.Printf("test_var: %s", testVar[0:100])
		os.Exit(0)
	}
	err := updateSecretsInStore(changedFiles, apiToken, apiUser)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func updateSecretsInStore(changedFiles, apiToken, apiUser string) error {
	changedFilesScanner := bufio.NewScanner(strings.NewReader(changedFiles))
	for changedFilesScanner.Scan() {
		syncSecretEvent, err := NewSecretUpdateEventFromFile(changedFilesScanner.Text())
		if err != nil {
			return fmt.Errorf("failed read sync secret event from file, error: %w", err)
		}

		secret, secretData, err := syncSecretEvent.getGCPSecret(changedFilesScanner.Text())
		if err != nil {
			return fmt.Errorf("failed get secret from GCP secret manager, error: %w", err)
		}

		fmt.Printf("Annotations: %s\n", secret.Annotations["jenkins-type"])

		creds, err := syncSecretEvent.newCredential(secret.Annotations)
		err = syncSecretEvent.getJenkinsCredentials(httpClient, secret, apiUser, apiToken, creds)
		if err != nil {
			return fmt.Errorf("failed get credentials from jenkins store, error: %w", err)
		}

		err = creds.UpdateSecret(secretData)
		if err != nil {
			return err
		}

		err = syncSecretEvent.updateJenkinsCredentials(httpClient, secret, apiUser, apiToken, creds)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("no changed files provided")
}

func NewSecretUpdateEventFromFile(syncSecretEventFilePath string) (SyncSecretEvent, error) {
	var syncSecretEvent SyncSecretEvent
	syncSecretEventFile, err := ioutil.ReadFile(syncSecretEventFilePath)
	if err != nil {
		return SyncSecretEvent{}, fmt.Errorf("failed read sync secret event file, error: %w", err)
	}
	err = yaml.Unmarshal(syncSecretEventFile, &syncSecretEvent)
	if err != nil {
		return SyncSecretEvent{}, fmt.Errorf("failed unmarshal yaml sync secret event file to struct, error: %w", err)
	}
	return syncSecretEvent, nil
}

func (sse *SyncSecretEvent) getGCPSecret(updateSecretEventFilePath string) (*gcpsecretmanager.Secret, string, error) {
	// TODO: secret path should be a field of SyncSecretEvent.
	tmpPath := strings.TrimPrefix(updateSecretEventFilePath, "/")
	gcpSecretManagerPath := strings.TrimSuffix(tmpPath, ".yaml")
	secret, err := secretManagerService.GetSecret(gcpSecretManagerPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed get secret from secret manager, error: %w", err)
	}
	// TODO: use path from SyncSecretEvent
	secretData, err := secretManagerService.GetLatestSecretVersionData(gcpSecretManagerPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed get latest version data, error: %w", err)
	}
	return secret, secretData, nil

}

// TODO: enhance sync event with required data for this calls
func (sse *SyncSecretEvent) getJenkinsCredentials(httpClient *http.Client, secret *gcpsecretmanager.Secret, apiUser, apiToken string, credential CredentialInterface) error {
	fmt.Printf("Secret Name: %s\n", secret.Name)
	shortSecretName := path.Base(secret.Name)
	// TODO: construct URL from vars and const
	req, err := http.NewRequest(http.MethodGet, "https://kymasecurity.jaas-gcp.cloud.sap.corp/job/administration/credentials/store/folder/domain/_/credential/"+shortSecretName+"/config.xml", http.NoBody)
	if err != nil {
		return fmt.Errorf("failed create http request, error: %w", err)
	}

	req.SetBasicAuth(apiUser, apiToken)

	response, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed calling jenkins api, error: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("calling jenkins api response status NOK, error: %w", err)
	}

	defer response.Body.Close()

	resBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed read http response body, error: %w", err)
	}

	err = xml.Unmarshal(resBody, &credential)
	if err != nil {
		return fmt.Errorf("failed unmarshal xml http response body to struct, error: %w", err)
	}
	return nil
}

func (sse *SyncSecretEvent) newCredential(secretAnnotations map[string]string) (CredentialInterface, error) {
	if secretAnnotations["jenkins-type"] == "secret-text" {
		return &SecretTextCredential{}, nil
	}
	return &Credential{}, fmt.Errorf("unknown credential type")
}

func (sse *SyncSecretEvent) updateJenkinsCredentials(httpClient *http.Client, secret *gcpsecretmanager.Secret, apiUser, apiToken string, credential CredentialInterface) error {
	bodyBytes, err := xml.Marshal(credential)
	if err != nil {
		return err
	}
	shortSecretName := path.Base(secret.Name)
	bodyReader := bytes.NewReader(bodyBytes)
	req, err := http.NewRequest(http.MethodPost, "https://kymasecurity.jaas-gcp.cloud.sap.corp/job/administration/credentials/store/folder/domain/_/credential/"+shortSecretName+"/config.xml", bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/xml")
	req.SetBasicAuth(apiUser, apiToken)

	response, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("calling jenkins api response status NOK, error: %w", err)
	}

	defer response.Body.Close()
	return nil
}
