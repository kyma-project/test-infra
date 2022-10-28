package sign

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

const (
	tagDelim         = ":"
	digestDelim      = "@"
	regRepoDelimiter = "/"
)

type ErrBackendNotSupported struct {
	Type string
}

func (e ErrBackendNotSupported) Error() string {
	return fmt.Sprintf("'%s' backend not supported", e.Type)
}

type SignerConfig struct {
	// Name contains the custom name of defined signer
	Name string `yaml:"name" json:"name"`
	// Type defines the type of signing backend.
	// Config will be parsed based on this value.
	Type string `yaml:"type" json:"type"`
	// Config defines specific configuration for signing backend.
	Config SignerFactory `yaml:"config" json:"config"`
}

type SignerFactory interface {
	NewSigner() (Signer, error)
}

type Signer interface {
	Sign([]string) error
}

func (sc *SignerConfig) UnmarshalYAML(value *yaml.Node) error {
	var t struct {
		Name string `yaml:"name"`
		Type string `yaml:"type"`
	}
	if err := value.Decode(&t); err != nil {
		return err
	}
	switch t.Type {
	case TypeNotaryBackend:
		var c struct {
			Config NotaryConfig `yaml:"config"`
		}
		if err := value.Decode(&c); err != nil {
			return err
		}
		sc.Config = c.Config
	default:
		return ErrBackendNotSupported{Type: t.Type}
	}
	sc.Type = t.Type
	sc.Name = t.Name
	return nil
}
