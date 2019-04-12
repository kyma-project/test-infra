package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/go-yaml/yaml"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

// Config holds application configuration
type Config struct {
	KymaProjectDir string `envconfig:"KYMA_PROJECT_DIR"`
	ClientID       string `envconfig:"GITHUB_INTEGRATION_APP_CLIENT_ID"`
	ClientSecret   string `envconfig:"GITHUB_INTEGRATION_APP_CLIENT_SECRET"`
	DexCallbackURL string `envconfig:"DEX_CALLBACK_URL"`
	KymaAdmins     string `envconfig:"GITHUB_TEAMS_WITH_KYMA_ADMINS_RIGHTS"`
}

const (
	clusterUsers = "/resources/core/charts/cluster-users/values.yaml"
	dexConfigMap = "/resources/dex/templates/dex-config-map.yaml"
)

func main() {
	var cfg Config
	err := envconfig.Init(&cfg)
	fatalOnError(err, "while init config")

	kymaPath := fmt.Sprintf("%s/kyma", cfg.KymaProjectDir)
	adjustClusterUsersOrDie(cfg, kymaPath)
	adjustDexConfigMapOrDie(cfg, kymaPath)
}

type (
	rootClusterUsers struct {
		Bindings bindings `yaml:"bindings"`
	}

	bindings struct {
		KymaAdmin groups `yaml:"kymaAdmin"`
	}

	groups struct {
		Groups []string `yaml:"groups"`
	}

	rawYAML map[interface{}]interface{}
)

func adjustClusterUsersOrDie(cfg Config, kymaPath string) {
	valuesPath := path.Join(kymaPath, clusterUsers)
	usersFile, err := ioutil.ReadFile(valuesPath)
	fatalOnError(err, "while reading file")

	cUsers := rootClusterUsers{}
	err = yaml.Unmarshal(usersFile, &cUsers)
	fatalOnError(err, "while unmarshaling file [typed]")

	teams := strings.Split(cfg.KymaAdmins, ",")
	for _, team := range teams {
		cUsers.Bindings.KymaAdmin.Groups = append(cUsers.Bindings.KymaAdmin.Groups, fmt.Sprintf("kyma-project:%s", team))
	}

	rawValues := rawYAML{}
	err = yaml.Unmarshal(usersFile, &rawValues)
	fatalOnError(err, "while unmarshaling file [raw]")

	rawValues["bindings"].(rawYAML)["kymaAdmin"].(rawYAML)["groups"] = cUsers.Bindings.KymaAdmin.Groups
	updated, err := yaml.Marshal(rawValues)
	fatalOnError(err, "while marshaling")

	err = ioutil.WriteFile(valuesPath, updated, os.ModeAppend)
	fatalOnError(err, "while saving file")
}

func adjustDexConfigMapOrDie(cfg Config, kymaPath string) {
	const githubConnectorPattern = `
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
	fConfigMap, err := os.OpenFile(path.Join(kymaPath, dexConfigMap), os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	fatalOnError(err, "while opening file")

	_, err = fConfigMap.Write([]byte(
		fmt.Sprintf(githubConnectorPattern, cfg.ClientID, cfg.ClientSecret, cfg.DexCallbackURL)))
	fatalOnError(err, "while appending github connector")

	err = fConfigMap.Close()
	fatalOnError(err, "while closing file")
}

func fatalOnError(err error, context string) {
	if err != nil {
		log.Fatal(errors.Wrap(err, context))
	}
}
