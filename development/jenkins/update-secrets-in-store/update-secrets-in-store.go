package main

import (
	"bufio"
	"context"
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
<org.jenkinsci.plugins.plaincredentials.impl.FileCredentialsImpl plugin="plain-credentials@139.ved2b_9cf7587b">
	<id>secret-manager-kymasecurity-jaas_sap-kyma-prow</id>
	<description></description>
	<fileName>sap-kyma-prow-03cf1cccd8eb.json</fileName>
	<secretBytes>
		<secret-redacted/>
	</secretBytes>
</org.jenkinsci.plugins.plaincredentials.impl.FileCredentialsImpl>
*/

var (
	secretManagerService *secretmanager.Service
	err                  error
	httpClient           *http.Client
)

type Credential struct {
	XMLName     xml.Name    `xml:"org.jenkinsci.plugins.plaincredentials.impl.FileCredentialsImpl"`
	Plugin      string      `xml:"plugin,attr"`
	ID          string      `xml:"id"`
	Description string      `xml:"description"`
	FileName    string      `xml:"fileName"`
	SecretBytes SecretBytes `xml:"secretBytes"`
}

type SecretBytes struct {
	XMLName        xml.Name `xml:"secretBytes"`
	SecretRedacted string   `xml:"secret-redacted"`
}

type SyncSecretEvent struct {
	SecretName      string   `yaml:"secret.name,omitempty"`
	SecretVersion   int      `yaml:"secret.version,omitempty"`
	SecretEndpoints []string `yaml:"secret.endpoints,omitempty"`
}

func main() {
	ctx := context.Background()
	secretManagerCreds := os.Getenv("SECRET_MANAGER_JSON")
	chanedFiles := os.Getenv("CHANGED_FILES")
	apiToken := os.Getenv("API_TOKEN")
	apiUser := os.Getenv("API_USER")
	credsOption := option.WithCredentialsJSON([]byte(secretManagerCreds))
	secretManagerService, err = secretmanager.NewService(ctx, credsOption)
	if err != nil {
		panic("failed creating Secret Manager client, error: " + err.Error())
	}
	httpClient = &http.Client{Timeout: 15 * time.Second}
	err := updateSecretsInStore(chanedFiles, apiToken, apiUser)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func updateSecretsInStore(changedFiles, apiToken, apiUser string) error {
	changedFilesScanner := bufio.NewScanner(strings.NewReader(changedFiles))
	for changedFilesScanner.Scan() {
		syncSecretEvent, err := NewSecretUpdateEventFromFile(changedFilesScanner.Text())
		if err != nil {
			return err
		}
		secret, secretData, err := syncSecretEvent.getGCPSecret(changedFilesScanner.Text())
		if err != nil {
			return err
		}
		creds, err := syncSecretEvent.getCredentialsFromStore(httpClient, secret, apiUser, apiToken)
		fmt.Printf("%+v", creds)
		fmt.Printf("%+v", secretData)
		return nil
	}
	return fmt.Errorf("no changed files provided")
}

func NewSecretUpdateEventFromFile(syncSecretEventFilePath string) (SyncSecretEvent, error) {
	var syncSecretEvent SyncSecretEvent
	syncSecretEventFile, err := ioutil.ReadFile(syncSecretEventFilePath)
	if err != nil {
		return SyncSecretEvent{}, err
	}
	err = yaml.Unmarshal(syncSecretEventFile, &syncSecretEvent)
	if err != nil {
		return SyncSecretEvent{}, err
	}
	return syncSecretEvent, nil
}

// TODO: this should be a method of secretmanagerservice
func (sse *SyncSecretEvent) getGCPSecret(updateSecretEventFilePath string) (*gcpsecretmanager.Secret, string, error) {
	// TODO: secret path should be a field of SyncSecretEvent.
	tmpPath := strings.TrimPrefix(updateSecretEventFilePath, "/")
	gcpSecretManagerPath := strings.TrimSuffix(tmpPath, ".yaml")
	secretName := path.Base(gcpSecretManagerPath)
	// Secrets filtering https://cloud.google.com/secret-manager/docs/filtering
	// TODO: How to build this filter. Should cloud run function get secret details and propagate some attributes and labels to github?
	secretsFilter := "name:" + secretName
	secrets, err := secretManagerService.GetAllSecrets(gcpSecretManagerPath, secretsFilter)
	if err != nil {
		fmt.Println("failed get secret from secret manager")
	}
	if len(secrets) > 1 {
		return nil, "", fmt.Errorf("to many secrets found")
	}
	// TODO: use path from SyncSecretEvent
	secretData, err := secretManagerService.GetLatestSecretVersionData(gcpSecretManagerPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed get latest version data")
	}
	return secrets[0], secretData, nil
	// authentication: 'dekiel-test-token', responseHandle: 'NONE', url: 'https://kymasecurity.jaas-gcp.cloud.sap.corp/job/administration/credentials/store/folder/domain/_/credential/secret-manager-kymasecurity-jaas_sap-kyma-prow/config.xml

}

// TODO: enhance sync event with required data for this calls
func (sse *SyncSecretEvent) getCredentialsFromStore(httpClient *http.Client, secret *gcpsecretmanager.Secret, apiUser, apiToken string) (Credential, error) {
	var credential Credential
	// TODO: construct URL from vars and const
	req, err := http.NewRequest(http.MethodGet, "https://kymasecurity.jaas-gcp.cloud.sap.corp/job/administration/credentials/store/folder/domain/_/credential/"+secret.Name+"/config.xml", http.NoBody)
	if err != nil {
		return Credential{}, err
	}

	req.SetBasicAuth(apiUser, apiToken)

	res, err := httpClient.Do(req)
	if err != nil {
		return Credential{}, err
	}

	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return Credential{}, err
	}

	err = xml.Unmarshal(resBody, &credential)
	if err != nil {
		return Credential{}, err
	}

	fmt.Printf("Status: %d\n", res.StatusCode)
	fmt.Printf("Body: %s\n", string(resBody))

	return credential, nil
}

/*
func (sse *SyncSecretEvent) updateCredentialsInStore() error {

}

*/
