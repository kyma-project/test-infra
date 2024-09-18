package imagesync

// SyncDef stores synchronisation definition
type SyncDef struct {
	Images []Image
}

// Image stores image location
type Image struct {
	Source    string
	Tag       string `yaml:"tag,omitempty"`
	AMD64Only bool   `yaml:"amd64Only,omitempty"`
}
