package pkg

// SyncDef stores synchronisation definition
type SyncDef struct {
	TargetRepoPrefix string `yaml:"targetRepoPrefix"`
	Images           []Image
}

// Image stores image location
type Image struct {
	Source    string
	Tag       string `yaml:"tag,omitempty"`
	AMD64Only bool   `yaml:"amd64Only,omitempty"`
}
