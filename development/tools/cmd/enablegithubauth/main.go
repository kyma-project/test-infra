package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/ghodss/yaml"
)

const (
	envKymaProjectDir                      = "KYMA_PROJECT_DIR"
	envDexGithubIntegrationAppClientID     = "GITHUB_INTEGRATION_APP_CLIENT_ID"
	envDexGithubIntegrationAppClientSecret = "GITHUB_INTEGRATION_APP_CLIENT_SECRET"
	envDexCallbackURL                      = "DEX_CALLBACK_URL"
	envGithubTeamsWithKymaAdmins           = "GITHUB_TEAMS_WITH_KYMA_ADMINS_RIGHTS"
)

func main() {
	kymaProjectDirVal := os.Getenv(envKymaProjectDir)
	clientID := os.Getenv(envDexGithubIntegrationAppClientID)
	clientSecret := os.Getenv(envDexGithubIntegrationAppClientSecret)
	dexCallbackURL := os.Getenv(envDexCallbackURL)
	kymaAdmins := os.Getenv(envGithubTeamsWithKymaAdmins)

	if kymaProjectDirVal == "" {
		log.Fatalf("missing env: %s", envKymaProjectDir)
	}
	if clientID == "" {
		log.Fatalf("missing env: %s", envDexGithubIntegrationAppClientID)
	}
	if clientSecret == "" {
		log.Fatalf("missing env: %s", envDexGithubIntegrationAppClientSecret)
	}
	if dexCallbackURL == "" {
		log.Fatalf("missing env: %s", envDexCallbackURL)
	}
	if kymaAdmins == "" {
		log.Fatalf("missing env: %s", envGithubTeamsWithKymaAdmins)
	}

	kymaPath := fmt.Sprintf("%s/kyma", kymaProjectDirVal)
	clusterUsers := "/resources/core/charts/cluster-users/values.yaml"
	dexConfigMap := "/resources/dex/templates/dex-config-map.yaml"

	fUsers, err := os.OpenFile(path.Join(kymaPath, clusterUsers), os.O_RDWR, os.ModeAppend)
	if err != nil {
		log.Fatalf("cannot open file %s, %v", path.Join(kymaPath, clusterUsers), err)
	}

	defer func() {
		if err := fUsers.Close(); err != nil {
			panic(err)
		}
	}()

	rClusterUsers := RootClusterUsers{}
	b, err := ioutil.ReadAll(fUsers)
	if err != nil {
		panic(err)
	}
	if err := yaml.Unmarshal(b, &rClusterUsers); err != nil {
		panic(err)
	}

	teams := strings.Split(kymaAdmins, ",")
	for _, team := range teams {
		rClusterUsers.Bindings.KymaAdmin.Groups = append(rClusterUsers.Bindings.KymaAdmin.Groups, fmt.Sprintf("kyma-project:%s", team))
	}

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
		fmt.Sprintf(githubConnectorPattern, clientID, clientSecret, dexCallbackURL)))

	if err != nil {
		panic(err)
	}

	if err := fConfigMap.Close(); err != nil {
		panic(err)
	}

}

var githubConnectorPattern = `
    connectors:
    - type: github
      id: github
      name: GitHub
      config:
        clientID: %s
        clientSecret: %s
        redirectURI: %s
        orgs:
        - name: kyma-project
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
