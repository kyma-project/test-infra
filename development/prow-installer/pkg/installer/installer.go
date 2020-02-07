package installer

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

//Configuration for prow installer.
type Config struct {
	ClusterName       string            `yaml:"cluster_name"`
	Oauth             string            `yaml:"oauth"`
	Project           string            `yaml:"project"`
	Zone              string            `yaml:"zone"`
	Location          string            `yaml:"location"`
	BucketName        string            `yaml:"bucket_name"`
	KeyringName       string            `yaml:"keyring_name"`
	EncryptionKeyName string            `yaml:"encryption_key_name"`
	Kubeconfig        string            `yaml:"kubeconfig,omitempty"`
	Prefix            string            `yaml:"prefix,omitempty"`
	ServiceAccounts   []Account         `yaml:"serviceAccounts"`
	GenericSecrets    []GenericSecret   `yaml:"generics,flow,omitempty"`
	Labels            map[string]string `yaml:"labels"`
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
func (installerConfig *Config) ReadConfig(configFilePath string) error {
	log.Debug("Reading config from %s", configFilePath)
	if configFile, err := ioutil.ReadFile(configFilePath); err != nil {
		return fmt.Errorf("failed reading config file %w", err)
	} else if err := yaml.Unmarshal(configFile, &installerConfig); err != nil {
		return fmt.Errorf("error when unmarshalling yaml file: %w", err)
	}
	for i, account := range installerConfig.ServiceAccounts {
		//TODO: add validation of Type property of Account type.
		if installerConfig.Prefix != "" {
			installerConfig.ServiceAccounts[i].Name = fmt.Sprintf("%s-%s", installerConfig.Prefix, account.Name)
		}
	}
	return nil
}
