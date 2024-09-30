package sign

import (
	"gopkg.in/yaml.v3"
)

const (
	TypeNotaryBackend = "notary"
)

type SignerConfig struct {
	// Name contains the custom name of defined signer
	Name string `yaml:"name" json:"name"`
	// Type defines the type of signing backend.
	// Config will be parsed based on this value.
	Type string `yaml:"type" json:"type"`
	// Config defines specific configuration for signing backend.
	Config SignerFactory `yaml:"config" json:"config"`
	// JobType contains list of ProwJob types that should be supported.
	// If the list is empty, the signer is enabled for all job types.
	// Usable only in CI mode (CI=true)
	JobType []string `yaml:"job-type" json:"job-type"`
}

type SignerFactory interface {
	NewSigner() (Signer, error)
}

type Signer interface {
	Sign([]string) error
}

func (sc *SignerConfig) UnmarshalYAML(value *yaml.Node) error {
	var t struct {
		Name    string    `yaml:"name"`
		Type    string    `yaml:"type"`
		JobType []string  `yaml:"job-type"`
		Config  yaml.Node `yaml:"config"`
	}
	if err := value.Decode(&t); err != nil {
		return err
	}

	sc.Type = t.Type
	sc.Name = t.Name
	sc.JobType = t.JobType

	var notaryConfig NotaryConfig
	if err := t.Config.Decode(&notaryConfig); err != nil {
		return err
	}
	sc.Config = &notaryConfig

	return nil
}
