package sign

import (
	"gopkg.in/yaml.v3"
	"testing"
)

func TestSignerConfig_UnmarshalYAML(t *testing.T) {
	tc := []struct {
		name       string
		expectType string
		expectName string
		expectErr  bool
		config     string
	}{
		{
			name:       "signer config type notary",
			expectName: "notary-config",
			expectType: "notary",
			config: `
name: notary-config
type: notary
config:
  endpoint: http://sign
  timeout: 10m`,
		},
		{
			name:      "backend not supported",
			expectErr: true,
			config: `
name: unknown-backend
type: unsupported
config:
  unsupported: true`,
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			var sc SignerConfig
			err := yaml.Unmarshal([]byte(c.config), &sc)
			if err != nil && !c.expectErr {
				t.Errorf("got error, but didn't want to: %v", err)
			}
			if c.expectName != sc.Name {
				t.Errorf("unmarshal wrong name: %v != %v", sc.Name, c.expectName)
			}
			if c.expectType != sc.Type {
				t.Errorf("unmarshal wrong type: %v != %v", sc.Type, c.expectType)
			}
			switch v := sc.Config.(type) {
			case NotaryConfig:
				if c.expectType != TypeNotaryBackend {
					t.Errorf("got wrong config type: %v", v)
				}
			}
		})
	}
}
