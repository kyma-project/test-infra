package config

import (
	"fmt"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/cluster"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/k8s"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/serviceaccount"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/storage"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

//Configuration for prow config.
type Config struct {
	Project           string                          `yaml:"project"`
	Region            string                          `yaml:"region"`
	Buckets           []storage.Bucket                `yaml:"buckets"`
	KeyringName       string                          `yaml:"keyring_name"`
	EncryptionKeyName string                          `yaml:"encryption_key_name"`
	Kubeconfig        string                          `yaml:"kubeconfig,omitempty"`
	Prefix            string                          `yaml:"prefix,omitempty"`
	ServiceAccounts   []serviceaccount.ServiceAccount `yaml:"serviceAccounts"`
	GenericSecrets    []k8s.GenericSecret             `yaml:"generics,flow,omitempty"`
	Labels            map[string]string               `yaml:"labels"`
	Clusters          map[string]cluster.Cluster      `yaml:"clusters"`
	GCSserviceAccount string                          `yaml:"gcs_serviceaccount"`
	GCSlogBucket      string                          `yaml:"gcs_log_bucket"`
}

//Get config configuration from yaml file.
func ReadConfig(configFilePath string) (*Config, error) {
	log.Debugf("Reading config from %s", configFilePath)
	var installerConfig Config
	if configFile, err := ioutil.ReadFile(configFilePath); err != nil {
		return nil, fmt.Errorf("failed reading config file %w", err)
	} else if err := yaml.Unmarshal(configFile, &installerConfig); err != nil {
		return nil, fmt.Errorf("error when unmarshalling yaml file: %w", err)
	}
	return &installerConfig, nil
}
