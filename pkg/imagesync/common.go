package imagesync

// SyncDef stores synchronisation definition
type SyncDef struct {
	Images []Image
}

// Image stores image location
type Image struct {
	Source string
	Tag    string `yaml:"tag,omitempty"`
}
