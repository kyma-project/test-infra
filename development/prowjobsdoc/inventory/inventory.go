package inventory

type Inventory struct {
	Total     int                    `yaml:"total"`
	Jobs      map[string]OrgRepoJobs `yaml:"jobs"`
	Periodics *Periodics             `yaml:"periodics,omitempty"`
}

type OrgRepoJobs struct {
	Total       int
	Presubmits  []OrgRepoJob `yaml:"presubmits,omitempty"`
	Postsubmits []OrgRepoJob `yaml:"postsubmits,omitempty"`
}

type OrgRepoJob struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	ExtraRefs   []Refs   `yaml:"extra_refs,omitempty"`
	BaseRef     *Refs    `yaml:"base_ref,omitempty"`
	Branches    []string `yaml:"branches,omitempty"`
}

type Periodics struct {
	Total int          `yaml:"total"`
	Jobs  []OrgRepoJob `yaml:"jobs,omitempty"`
}

type Refs struct {
	BaseRef string `yaml:"base_ref"`
	Repo    string `yaml:"repo"`
}
