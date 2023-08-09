package main

import (
	"testing"

	"github.com/kyma-project/test-infra/development/types"
)

const (
	invalidUsername    = "invalid"
	githubUsername     = "github"
	enterpriseUsername = "enterprise"
	slackUsername      = "slack"

	githubDomain     = "github.com"
	enterpriseDomain = "github.tools.sap"
)

func TestGetSlackUsername(t *testing.T) {
	usersMap := []types.User{
		{
			ComGithubUsername:          githubUsername,
			SapToolsGithubUsername:     enterpriseUsername,
			ComEnterpriseSlackUsername: slackUsername,
		},
	}
	tests := []struct {
		name                  string
		expectedSlackUsername string
		domain                string
		githubUsername        string
	}{
		{
			name:                  "Existing user in github domain",
			expectedSlackUsername: slackUsername,
			domain:                githubDomain,
			githubUsername:        githubUsername,
		},
		{
			name:                  "Existing user in enterprise domain",
			expectedSlackUsername: slackUsername,
			domain:                enterpriseDomain,
			githubUsername:        enterpriseUsername,
		},
		{
			name:                  "nonexisting user in github domain",
			expectedSlackUsername: "",
			domain:                githubDomain,
			githubUsername:        invalidUsername,
		},
		{
			name:                  "nonexisting user in nonexisting domain",
			expectedSlackUsername: "",
			domain:                invalidUsername,
			githubUsername:        invalidUsername,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			slackUsername := getSlackUsername(usersMap, test.githubUsername, test.domain)
			if slackUsername != test.expectedSlackUsername {

			}
		})
	}
}
