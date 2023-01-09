package types

import (
	"github.com/zricethezav/gitleaks/v8/report"
)

// User holds kyma development team user details.
// It provides mapping of various details used for integration different systems.
// It holds information about automerge notification preferences.
type User struct {
	ComGithubUsername          string `yaml:"com.github.username,omitempty"`
	SapToolsGithubUsername     string `yaml:"sap.tools.github.username,omitempty"`
	ComEnterpriseSlackUsername string `yaml:"com.slack.enterprise.sap.username,omitempty"`
	AutomergeNotifications     bool   `yaml:"automerge.notification,omitempty"`
}

// Alias holds mapping between owners file alias and slack groups and channels names.
// It holds information if automerge notification is enabled.
type Alias struct {
	ComGithubAliasname              string   `yaml:"com.github.aliasname,omitempty"`
	ComEnterpriseSlackGroupsnames   []string `yaml:"com.slack.enterprise.sap.groupsnames,omitempty"`
	ComEnterpriseSlackChannelsnames []string `yaml:"com.slack.enterprise.sap.channelsnames,omitempty"`
	AutomergeNotifications          bool     `yaml:"automerge.notification,omitempty"`
}

// TODO: Migrate all usages to development/gcp/pkg/types package.
type GCPStorageMetadata struct {
	BucketName *string `json:"bucketName,omitempty"`
	Directory  *string `json:"directory,omitempty"`
}

// TODO: Migrate all usages to development/gcp/pkg/types package.
type GCPProjectMetadata struct {
	Project *string `json:"project,omitempty"`
}

// TODO: Migrate all usages to development/github/pkg/types package.
type GithubIssueMetadata struct {
	GithubOrg         *string `json:"githubOrg,omitempty"`
	GithubRepo        *string `json:"githubRepo,omitempty"`
	GithubIssueNumber *int    `json:"githubIssueNumber,omitempty"`
	GithubIssueURL    *string `json:"githubIssueURL,omitempty"`
	GithubAssignee    User    `json:"githubIssueAssignee,omitempty"`
}

type SecretsLeakScannerMessage struct {
	LeaksFound  *bool            `json:"leaksFound"`
	LeaksReport []report.Finding `json:"leaksReport,omitempty"`
}
