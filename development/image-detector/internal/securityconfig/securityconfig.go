package securityconfig

import (
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Whitesource struct {
	Language    string   `yaml:"language"`
	SubProjects bool     `yaml:"subprojects"`
	Exclude     []string `yaml:"exclude"`
}

type SecurityConfig struct {
	ModuleName  string      `yaml:"module-name"`
	Images      []string    `yaml:"protecode"`
	Whitesource Whitesource `yaml:"whitesource"`
}

func LoadSecurityConfig(path string) (*SecurityConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return parseSecurityConfig(f)
}

func parseSecurityConfig(reader io.Reader) (*SecurityConfig, error) {
	var securityConfig SecurityConfig
	err := yaml.NewDecoder(reader).Decode(&securityConfig)
	if err != nil {
		return nil, err
	}

	return &securityConfig, nil
}

func (config *SecurityConfig) SaveToFile(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	err = yaml.NewEncoder(f).Encode(config)
	if err != nil {
		return err
	}

	return nil
}
