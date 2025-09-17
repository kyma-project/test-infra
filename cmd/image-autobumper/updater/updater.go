/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package updater handles creation and updates of GitHub PullRequests.
package updater

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"sigs.k8s.io/prow/pkg/github"
)

// Constants to indicate whether maintainers can modify a pull request in a fork.
const (
	// AllowMods indicates that maintainers can modify the pull request.
	AllowMods = true

	// PreventMods indicates that maintainers cannot modify the pull request.
	PreventMods = false
)

// Query constants used in the GitHub search query.
const (
	// queryState indicates the state of the pull requests to search for (open).
	queryState = "is:open"

	// queryType indicates the type of the items to search for (pull requests).
	queryType = "is:pr"

	// queryArchived excludes archived repositories from the search.
	queryArchived = "archived:false"

	// querySortField indicates the field to sort the search results by (updated time).
	querySortField = "updated"

	// firstIssueIndex is the index of the first issue in the search results.
	firstIssueIndex = 0
)

type updateClient interface {
	UpdatePullRequest(org, repo string, number int, title, body *string, open *bool, branch *string, canModify *bool) error
	BotUser() (*github.UserData, error)
	FindIssues(query, sort string, asc bool) ([]github.Issue, error)
}

type ensureClient interface {
	updateClient
	AddLabel(org, repo string, number int, label string) error
	CreatePullRequest(org, repo, title, body, head, base string, canModify bool) (int, error)
	GetIssue(org, repo string, number int) (*github.Issue, error)
}

// EnsurePRWithQueryTokens ensures that a pull request exists with the given parameters.
// It reuses an existing pull request if one matches the query tokens, otherwise it creates a new one.
func EnsurePRWithQueryTokens(org, repo, title, body, source, baseBranch, queryTokensString string, allowMods bool, gc ensureClient) (*int, error) {
	prNumber, err := updatePRWithQueryTokens(org, repo, title, body, queryTokensString, gc)
	if err != nil {
		return nil, fmt.Errorf("update error: %w", err)
	}

	if prNumber == nil {
		pr, err := gc.CreatePullRequest(org, repo, title, body, source, baseBranch, allowMods)
		if err != nil {
			return nil, fmt.Errorf("create error: %w", err)
		}
		logrus.Infof("Created new PR with number: %d", pr)
		prNumber = &pr
	} else {
		logrus.Infof("Reused existing PR with number: %d", *prNumber)
	}

	return prNumber, nil
}

// updatePRWithQueryTokens looks for an existing PR to reuse based on the provided query tokens.
// If found, it updates the PR; otherwise, it returns nil.
func updatePRWithQueryTokens(org, repo, title, body, queryTokensString string, gc updateClient) (*int, error) {
	logrus.Info("Looking for a PR to reuse...")

	// Get the bot user
	me, err := gc.BotUser()
	if err != nil {
		return nil, fmt.Errorf("bot name: %w", err)
	}

	// Construct the query to find issues
	query := fmt.Sprintf("%s %s %s repo:%s/%s author:%s %s", queryState, queryType, queryArchived, org, repo, me.Login, queryTokensString)

	// Find issues based on the query
	issues, err := gc.FindIssues(query, querySortField, false)
	if err != nil {
		return nil, fmt.Errorf("find issues: %w", err)
	}

	// If no reusable issues are found, return nil
	if len(issues) == 0 {
		logrus.Info("No reusable issues found")
		return nil, nil
	}

	// Pick the first issue (most recently updated)
	prNumber := issues[firstIssueIndex].Number
	logrus.Infof("Found PR #%d", prNumber)

	// Prepare to ignore certain fields in the update request
	var ignoreOpen *bool
	var ignoreBranch *string
	var ignoreModify *bool

	// Update the pull request with the new title and body
	if err := gc.UpdatePullRequest(org, repo, prNumber, &title, &body, ignoreOpen, ignoreBranch, ignoreModify); err != nil {
		return nil, fmt.Errorf("update PR #%d: %w", prNumber, err)
	}

	return &prNumber, nil
}

func EnsurePRWithLabels(org, repo, title, body, source, baseBranch, headBranch string, allowMods bool, gc ensureClient, labels []string) (*int, error) {
	return EnsurePRWithQueryTokensAndLabels(org, repo, title, body, source, baseBranch, "head:"+headBranch, allowMods, labels, gc)
}

func EnsurePRWithQueryTokensAndLabels(org, repo, title, body, source, baseBranch, queryTokensString string, allowMods bool, labels []string, gc ensureClient) (*int, error) {
	n, err := EnsurePRWithQueryTokens(org, repo, title, body, source, baseBranch, queryTokensString, allowMods, gc)
	if err != nil {
		return n, err
	}

	if len(labels) == 0 {
		return n, nil
	}

	issue, err := gc.GetIssue(org, repo, *n)
	if err != nil {
		return n, fmt.Errorf("failed to get PR: %w", err)
	}

	for _, label := range labels {
		if issue.HasLabel(label) {
			continue
		}

		if err := gc.AddLabel(org, repo, *n, label); err != nil {
			return n, fmt.Errorf("failed to add label %q: %w", label, err)
		}
		logrus.WithField("label", label).Info("Added label")
	}
	return n, nil
}
