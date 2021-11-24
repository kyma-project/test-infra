package pkg

// SyncDef stores synchronisation definition
type SyncDef struct {
	TargetRepoPrefix string `yaml:"targetRepoPrefix"`
	Sign             bool   `yaml:"sign"`
	Images           []Image
}

// Image stores image location
type Image struct {
	Source string
	Tag    string `yaml:"tag,omitempty"`
	Sign   *bool  `yaml:"sign,omitempty"`
}
