package installer

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"prow_installer/accessmanager"
)

var (
	config          = flag.String("config", "", "Config file path [Required]")
	credentialsfile = flag.String("credentialsfile", "", "Google Application Credentials file path [Required]")
	prefix          = flag.String("prefix", "", "Prefix for naming resources [Optional]")
)

//Configuration for prow installer.
type installerConfig struct {
	ClusterName       string          `yaml:"cluster_name"`
	Oauth             string          `yaml:"oauth"`
	Project           string          `yaml:"project"`
	Zone              string          `yaml:"zone"`
	Location          string          `yaml:"location"`
	BucketName        string          `yaml:"bucket_name"`
	KeyringName       string          `yaml:"keyring_name"`
	EncryptionKeyName string          `yaml:"encryption_key_name"`
	Kubeconfig        string          `yaml:"kubeconfig,omitempty"`
	ServiceAccounts   serviceAccounts `yaml:"serviceAccounts"`
	GenericSecrets    genericSecrets  `yaml:"generics,flow,omitempty"`
}

type serviceAccounts []serviceAccount

type serviceAccount struct {
	Name  string   `yaml:"name"`
	Roles []string `yaml:"roles"`
}

type genericSecrets []genericSecret

type genericSecret struct {
	Name string `yaml:"prefix"`
	Key  string `yaml:"key"`
}

//Get installer configuration from yaml file.
func getInstallerConfig(configFilePath string, config *installerConfig) *installerConfig {
	configfile, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Printf("Error %v when reading file %s", err, configFilePath)
	}
	err = yaml.Unmarshal(configfile, config)
	if err != nil {
		log.Fatalf("Error %v when unmarshalling yaml file.", err)
	}
	return config
}
