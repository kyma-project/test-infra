package imagelister

type CustomFields struct {
	Components string `json:"components,omitempty"`
	Image      string `json:"image"`
}
type ImageJSON struct {
	Name         string       `json:"name"`
	CustomFields CustomFields `json:"custom_fields"`
}

type ImagesJSON struct {
	Images []ImageJSON `json:"images"`
}
