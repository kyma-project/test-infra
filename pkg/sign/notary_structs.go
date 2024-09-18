package sign

import "time"

type Target struct {
	Name     string `json:"name"`
	ByteSize int64  `json:"byteSize"`
	Digest   string `json:"digest"`
}

type TrustedCollection struct {
	GUN     string   `json:"gun"`
	Targets []Target `json:"targets"`
}

type SigningPayload struct {
	TrustedCollections []TrustedCollection `json:"trustedCollections"`
}

type SignifySecret struct {
	CertificateData string `json:"certData"`
	PrivateKeyData  string `json:"privateKeyData"`
}

// NotaryConfig structs
type NotaryConfig struct {
	Endpoint     string            `yaml:"endpoint" json:"endpoint"`
	Secret       *AuthSecretConfig `yaml:"secret,omitempty" json:"secret,omitempty"`
	Timeout      time.Duration     `yaml:"timeout" json:"timeout"`
	RetryTimeout time.Duration     `yaml:"retry-timeout" json:"retry-timeout"`
	ReadFileFunc func(string) ([]byte, error)
}

type AuthSecretConfig struct {
	Path string `yaml:"path" json:"path"`
	Type string `yaml:"type" json:"type"`
}
