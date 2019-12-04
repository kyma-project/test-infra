package installer

import (
	"fmt"
	"io/ioutil"

	"github.com/kyma-project/test-infra/development/prow-installer/pkg/gcpaccessmanager"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

//Configuration for prow installer.
type Config struct {
	ClusterName       string `yaml:"cluster_name"`
	Oauth             string `yaml:"oauth"`
	Project           string `yaml:"project"`
	Zone              string `yaml:"zone"`
	Location          string `yaml:"location"`
	BucketName        string `yaml:"bucket_name"`
	KeyringName       string `yaml:"keyring_name"`
	EncryptionKeyName string `yaml:"encryption_key_name"`
	Kubeconfig        string `yaml:"kubeconfig,omitempty"`
	Prefix            string `yaml:"prefix,omitempty"`
	// TODO: Design objects for gke and azure different account types. Do you need another module, where to declare them, how to provide data in config.
	// At present there are two account types, gke service account, gke user.
	ServiceAccounts []gcpaccessmanager.Account `yaml:"serviceAccounts"`
	GenericSecrets  []GenericSecret            `yaml:"generics,flow,omitempty"`
}

type GenericSecret struct {
	Name string `yaml:"prefix"`
	Key  string `yaml:"key"`
}

//Get installer configuration from yaml file.
func (installerconfig *Config) ReadConfig(configFilePath string) {
	configfile, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatalf("Error %v when reading file %s", err, configFilePath)
	}
	err = yaml.Unmarshal(configfile, &installerconfig)
	if err != nil {
		log.Fatalf("Error %v when unmarshalling yaml file.", err)
	}
	for i, account := range installerconfig.ServiceAccounts {
		//TODO: add validation of Type property of Account type.
		if installerconfig.Prefix != "" {
			installerconfig.ServiceAccounts[i].Name = fmt.Sprintf("%s-%s", installerconfig.Prefix, account.Name)
		}
	}
}
