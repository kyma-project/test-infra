package config

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type LogsScrapingConfig struct {
	ChannelID         string   `yaml:"channelID"`
	ChannelName       string   `yaml:"channelName"`
	TestCases         []string `yaml:"testCases"`
	OnlyReportFailure bool     `yaml:"onlyReportFailure"`
}

type Dispatching struct {
	Config []LogsScrapingConfig
}

func LoadDispatchingConfig(path string) (Dispatching, error) {
	dispatchingConfig := []LogsScrapingConfig{}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return Dispatching{}, errors.Wrapf(err, "while reading configuration from %s", path)
	}

	if err := yaml.Unmarshal(content, &dispatchingConfig); err != nil {
		return Dispatching{}, errors.Wrapf(err, "while unmarshalling configuration from %s", path)
	}

	return Dispatching{
		Config: dispatchingConfig,
	}, nil
}

func contains(slice []string, element string) bool {
	// yeah, why create such a function in stdlib, who would need it? /s
	for _, s := range slice {
		if s == element {
			return true
		}
	}
	return false
}

func (d Dispatching) GetConfigByName(name string) (LogsScrapingConfig, error) {
	for _, conf := range d.Config {
		if contains(conf.TestCases, name) {
			return conf, nil
		}
	}

	return LogsScrapingConfig{}, fmt.Errorf("there's no configuration for %s test case", name)
}

func (d Dispatching) GetConfigByNameWithFallback(name string) (LogsScrapingConfig, error) {
	config, err := d.GetConfigByName(name)
	if err == nil {
		return config, err
	}

	return d.GetConfigByName("default")
}

func (d Dispatching) Validate() error {
	for _, config := range d.Config {
		if !strings.HasPrefix(config.ChannelName, "#") {
			return fmt.Errorf("channelName %s should start with #", config.ChannelName)
		}
	}
	return nil
}
