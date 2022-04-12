package types

// User holds kyma development team user details.
// It provides mapping of various details used for integration different systems.
// It holds information about automerge notification preferences.
type User struct {
	ComGithubUsername          string `yaml:"com.github.username,omitempty"`
	SapToolsGithubUsername     string `yaml:"sap.tools.github.username,omitempty"`
	ComEnterpriseSlackUsername string `yaml:"com.slack.enterprise.sap.username,omitempty"`
	AutomergeNotifications     bool   `yaml:"automerge.notification,omitempty"`
}

// Alias holds mapping betwen owners file alias and slack groups and channels names.
// It holds information if automerge notification is enabled.
type Alias struct {
	ComGithubAliasname              string   `yaml:"com.github.aliasname,omitempty"`
	ComEnterpriseSlackGroupsnames   []string `yaml:"com.slack.enterprise.sap.groupsnames,omitempty"`
	ComEnterpriseSlackChannelsnames []string `yaml:"com.slack.enterprise.sap.channelsnames,omitempty"`
	AutomergeNotifications          bool     `yaml:"automerge.notification,omitempty"`
}
