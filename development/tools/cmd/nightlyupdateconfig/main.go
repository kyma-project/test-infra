package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/ghodss/yaml"
)

const (
	envKymaProjectDir                      = "KYMA_PROJECT_DIR"
	envDexGithubIntegrationAppClientID     = "DEX_GITHUB_INTEGRATION_APP_CLIENT_ID"
	envDexGithubIntegrationAppClientSecret = "DEX_GITHUB_INTEGRATION_APP_CLIENT_SECRET"
)

func main() {
	kymaProjectDirVal := os.Getenv(envKymaProjectDir)
	clientID := os.Getenv(envDexGithubIntegrationAppClientID)
	clientSecret := os.Getenv(envDexGithubIntegrationAppClientSecret)
	if kymaProjectDirVal == "" {
		log.Fatalf("missing env: %s", envKymaProjectDir)
	}
	if clientID == "" {
		log.Fatalf("missing env: %s", envDexGithubIntegrationAppClientID)
	}
	if clientSecret == "" {
		log.Fatalf("missing env: %s", envDexGithubIntegrationAppClientSecret)
	}

	kymaPath := fmt.Sprintf("%s/kyma", kymaProjectDirVal)
	clusterUsers := "/resources/core/charts/cluster-users/values.yaml"
	dexConfigMap := "/resources/dex/templates/dex-config-map.yaml"

	fUsers, err := os.OpenFile(path.Join(kymaPath, clusterUsers), os.O_RDWR, os.ModeAppend)
	if err != nil {
		panic(err)
	}

	defer fUsers.Close()

	rClusterUsers := RootClusterUsers{}
	b, err := ioutil.ReadAll(fUsers)
	if err != nil {
		panic(err)
	}
	if err := yaml.Unmarshal(b, &rClusterUsers); err != nil {
		panic(err)
	}

	rClusterUsers.Bindings.KymaAdmin.Groups = append(rClusterUsers.Bindings.KymaAdmin.Groups, "aszecowka-org:only-adam-team")
	n, err := yaml.Marshal(rClusterUsers)
	if err != nil {
		panic(err)
	}

	_, err = fUsers.Seek(0, 0)
	if err != nil {
		panic(err)
	}
	_, err = fUsers.Write(n)
	if err != nil {
		panic(err)
	}

	fConfigMap, err := os.OpenFile(path.Join(kymaPath, dexConfigMap), os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		panic(err)
	}

	_, err = fConfigMap.Write([]byte(
		fmt.Sprintf(githubConnectorPattern, clientID, clientSecret)))

	if err != nil {
		panic(err)
	}

	fConfigMap.Close()

}

var githubConnectorPattern = `
    connectors:
    - type: github
      id: github
      name: GitHub
      config:
        clientID: %s
        clientSecret: %s
        redirectURI: https://dex.kyma.local/callback
        orgs:
        - name: aszecowka-org
`

// RootClusterUsers .
type RootClusterUsers struct {
	Bindings Bindings `json:"bindings"`
}

// Bindings .
type Bindings struct {
	KymaAdmin Groups `json:"kymaAdmin"`
	KymaView  Groups `json:"kymaView"`
}

// Groups .
type Groups struct {
	Groups []string `json:"groups"`
}
