package main

import (
	"testing"

	"github.com/kyma-project/test-infra/pkg/types"
)

const (
	invalidUsername    = "invalid"
	githubUsername     = "github"
	enterpriseUsername = "enterprise"
	slackUsername      = "slack"
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
		githubUsername        string
	}{
		{
			name:                  "Existing user",
			expectedSlackUsername: slackUsername,
			githubUsername:        enterpriseUsername,
		},
		{
			name:                  "nonexisting user",
			expectedSlackUsername: "",
			githubUsername:        invalidUsername,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			slackUsername := getSlackUsername(usersMap, test.githubUsername)
			if slackUsername != test.expectedSlackUsername {

			}
		})
	}
}
