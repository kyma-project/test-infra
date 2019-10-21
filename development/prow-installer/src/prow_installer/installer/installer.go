package installer

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

//Configuration for prow installer.
type InstallerConfig struct {
	ClusterName       string          `yaml:"cluster_name"`
	Oauth             string          `yaml:"oauth"`
	Project           string          `yaml:"project"`
	Zone              string          `yaml:"zone"`
	Location          string          `yaml:"location"`
	BucketName        string          `yaml:"bucket_name"`
	KeyringName       string          `yaml:"keyring_name"`
	EncryptionKeyName string          `yaml:"encryption_key_name"`
	Kubeconfig        string          `yaml:"kubeconfig,omitempty"`
	Prefix            string          `yaml:"prefix,omitempty"`
	ServiceAccounts   []Account       `yaml:"serviceAccounts"`
	GenericSecrets    []GenericSecret `yaml:"generics,flow,omitempty"`
}

//type Accounts []Account

//TODO: Should this be moved to accessmanager package and imported here? As methods from accessmanager pacakge expect this type as argument.
type Account struct {
	Name  string   `yaml:"name"`
	Type  string   `yaml:"type"`
	Roles []string `yaml:"roles,omitempty"`
}

//type GenericSecrets []GenericSecret

type GenericSecret struct {
	Name string `yaml:"prefix"`
	Key  string `yaml:"key"`
}

//Get installer configuration from yaml file.
func (installerconfig *InstallerConfig) ReadConfig(configFilePath string) {
	configfile, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Printf("Error %v when reading file %s", err, configFilePath)
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
