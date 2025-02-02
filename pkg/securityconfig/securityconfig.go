package securityconfig

import (
	"io"
	"os"
	"reflect"

	"gopkg.in/yaml.v3"
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

// TODO(kacpermalachowski): Remove after migration to the new field names
// see: https://github.tools.sap/kyma/test-infra/issues/491
func (config *SecurityConfig) UnmarshalYAML(value *yaml.Node) error {
	// Cannot use inheritance due to infinite loop
	var cfg struct {
		ModuleName   string       `yaml:"module-name,omitempty"`
		RcTag        string       `yaml:"rc-tag,omitempty"`
		Kind         string       `yaml:"kind,omitempty"`
		Images       []string     `yaml:"bdba,omitempty"`
		Mend         Mend         `yaml:"mend,omitempty"`
		CheckmarxOne CheckmarxOne `yaml:"checkmarx-one,omitempty"`
		Protecode    []string     `yaml:"protecode,omitempty"`
		Whitesource  Mend         `yaml:"whitesource,omitempty"`
	}

	if err := value.Decode(&cfg); err != nil {
		return err
	}

	config.ModuleName = cfg.ModuleName
	config.RcTag = cfg.RcTag
	config.Kind = cfg.Kind
	config.Images = cfg.Images
	config.Mend = cfg.Mend

	if len(cfg.Protecode) > 0 {
		config.Images = cfg.Protecode
	}

	if !reflect.DeepEqual(cfg.Whitesource, Mend{}) {
		config.Mend = cfg.Whitesource
	}

	return nil
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
