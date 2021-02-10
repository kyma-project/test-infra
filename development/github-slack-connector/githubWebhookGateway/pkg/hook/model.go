package hook

//PayloadDetails defines all fields of github's POST method for creating webhooks
type PayloadDetails struct {
	Name   string   `json:"name"`
	Active bool     `json:"active"`
	Config Config   `json:"config"`
	Events []string `json:"events,omitempty"`
}

//Config defines the structure of HookJSON's config
type Config struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type,omitempty"`
	Secret      string `json:"secret,omitempty"`
	InsecureSSL string `json:"insecure_ssl,omitempty"`
}
