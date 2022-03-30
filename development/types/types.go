package types

// User holds kyma development team user details.
// It provides mapping of various details used for integration different systems.
type User struct {
	ComGithubUsername          string `yaml:"com.github.username,omitempty"`
	SapToolsGithubUsername     string `yaml:"sap.tools.github.username,omitempty"`
	ComEnterpriseSlackUsername string `yaml:"com.slack.enterprise.sap.username,omitempty"`
	SkipAutomergeNotifications bool   `yaml:"skip.automerge.notification,omitempty"`
}

type Alias struct {
	ComGithubAliasname              string   `yaml:"com.github.aliasname,omitempty"`
	ComEnterpriseSlackGroupsnames   []string `yaml:"com.slack.enterprise.sap.groupsnames,omitempty"`
	ComEnterpriseSlackChannelsnames []string `yaml:"com.slack.enterprise.sap.channelsnames,omitempty"`
	SkipAutomergeNotifications      bool     `yaml:"skip.automerge.notification,omitempty"`
}

// TODO: this should be moved to development/logging module

type Logger interface {
	LogCritical(string)
	LogError(string)
	LogInfo(string)
}
