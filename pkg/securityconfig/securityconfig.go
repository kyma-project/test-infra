package securityconfig

import (
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

type CheckmarxOne struct {
	Preset  string   `yaml:"preset,omitempty"`
	Exclude []string `yaml:"exclude,omitempty"`
}

type Mend struct {
	Language    string   `yaml:"language,omitempty"`
	SubProjects bool     `yaml:"subprojects,omitempty"`
	Exclude     []string `yaml:"exclude,omitempty"`
}

type SecurityConfig struct {
	ModuleName   string       `yaml:"module-name,omitempty"`
	RcTag        string       `yaml:"rc-tag,omitempty"`
	Kind         string       `yaml:"kind,omitempty"`
	Images       []string     `yaml:"bdba,omitempty"`
	Mend         Mend         `yaml:"mend,omitempty"`
	CheckmarxOne CheckmarxOne `yaml:"checkmarx-one,omitempty"`
}

func ParseSecurityConfig(reader io.Reader) (*SecurityConfig, error) {
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
