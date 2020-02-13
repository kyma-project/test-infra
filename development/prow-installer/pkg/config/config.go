package config

import (
	"fmt"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/cluster"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

//Configuration for prow config.
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
	Clusters          []cluster.Cluster `yaml:"clusters"`
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

//Get config configuration from yaml file.
func ReadConfig(configFilePath string) (*Config, error) {
	log.Debug("Reading config from %s", configFilePath)
	var installerConfig Config
	if configFile, err := ioutil.ReadFile(configFilePath); err != nil {
		return nil, fmt.Errorf("failed reading config file %w", err)
	} else if err := yaml.Unmarshal(configFile, &installerConfig); err != nil {
		return nil, fmt.Errorf("error when unmarshalling yaml file: %w", err)
	}
	return &installerConfig, nil
}
