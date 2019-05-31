/*
Copyright 2016 The Kubernetes Authors.

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

package trigger

import (
	"fmt"
	"regexp"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/labels"
	"k8s.io/test-infra/prow/plugins"
)

var okToTestRe = regexp.MustCompile(`(?m)^/ok-to-test\s*$`)
var testAllRe = regexp.MustCompile(`(?m)^/test all,?($|\s.*)`)
var retestRe = regexp.MustCompile(`(?m)^/retest\s*$`)

func handleGenericComment(c Client, trigger plugins.Trigger, gc github.GenericCommentEvent) error {
	org := gc.Repo.Owner.Login
	repo := gc.Repo.Name
	number := gc.Number
	commentAuthor := gc.User.Login
	// Only take action when a comment is first created,
	// when it belongs to a PR,
	// and the PR is open.
	if gc.Action != github.GenericCommentActionCreated || !gc.IsPR || gc.IssueState != "open" {
		return nil
	}
	// Skip comments not germane to this plugin
	if !retestRe.MatchString(gc.Body) && !okToTestRe.MatchString(gc.Body) && !testAllRe.MatchString(gc.Body) {
		matched := false
		for _, presubmit := range c.Config.Presubmits[gc.Repo.FullName] {
			matched = matched || presubmit.TriggerMatches(gc.Body)
			if matched {
				break
			}
		}
		if !matched {
			c.Logger.Debug("Comment doesn't match any triggering regex, skipping.")
			return nil
		}
	}

	// Skip bot comments.
	botName, err := c.GitHubClient.BotName()
	if err != nil {
		return err
	}
	if commentAuthor == botName {
		c.Logger.Debug("Comment is made by the bot, skipping.")
		return nil
	}

	pr, err := c.GitHubClient.GetPullRequest(org, repo, number)
	if err != nil {
		return err
	}

	// Skip untrusted users comments.
	trusted, err := TrustedUser(c.GitHubClient, trigger, commentAuthor, org, repo)
	if err != nil {
		return fmt.Errorf("error checking trust of %s: %v", commentAuthor, err)
	}
	var l []github.Label
	if !trusted {
		// Skip untrusted PRs.
		l, trusted, err = TrustedPullRequest(c.GitHubClient, trigger, gc.IssueAuthor.Login, org, repo, number, nil)
		if err != nil {
			return err
		}
		if !trusted {
			resp := fmt.Sprintf("Cannot trigger testing until a trusted user reviews the PR and leaves an `/ok-to-test` message.")
			c.Logger.Infof("Commenting \"%s\".", resp)
			return c.GitHubClient.CreateComment(org, repo, number, plugins.FormatResponseRaw(gc.Body, gc.HTMLURL, gc.User.Login, resp))
		}
	}

	// At this point we can trust the PR, so we eventually update labels.
	// Ensure we have labels before test, because TrustedPullRequest() won't be called
	// when commentAuthor is trusted.
	if l == nil {
		l, err = c.GitHubClient.GetIssueLabels(org, repo, number)
		if err != nil {
			return err
		}
	}
	isOkToTest := HonorOkToTest(trigger) && okToTestRe.MatchString(gc.Body)
	if isOkToTest && !github.HasLabel(labels.OkToTest, l) {
		if err := c.GitHubClient.AddLabel(org, repo, number, labels.OkToTest); err != nil {
			return err
		}
	}
	if (isOkToTest || github.HasLabel(labels.OkToTest, l)) && github.HasLabel(labels.NeedsOkToTest, l) {
		if err := c.GitHubClient.RemoveLabel(org, repo, number, labels.NeedsOkToTest); err != nil {
			return err
		}
	}

	toTest, toSkip, err := FilterPresubmits(HonorOkToTest(trigger), c.GitHubClient, gc.Body, pr, c.Config.Presubmits[gc.Repo.FullName], c.Logger)
	if err != nil {
		return err
	}
	return runAndSkipJobs(c, pr, toTest, toSkip, gc.GUID, trigger.ElideSkippedContexts)
}

func HonorOkToTest(trigger plugins.Trigger) bool {
	return !trigger.IgnoreOkToTest
}

type GitHubClient interface {
	GetCombinedStatus(org, repo, ref string) (*github.CombinedStatus, error)
	GetPullRequestChanges(org, repo string, number int) ([]github.PullRequestChange, error)
}

// FilterPresubmits determines which presubmits should run. We only want to
// trigger jobs that should run, but the pool of jobs we filter to those that
// should run depends on the type of trigger we just got:
//  - if we get a /test foo, we only want to consider those jobs that match;
//    jobs will default to run unless we can determine they shouldn't
//  - if we got a /retest, we only want to consider those jobs that have
//    already run and posted failing contexts to the PR or those jobs that
//    have not yet run but would otherwise match /test all; jobs will default
//    to run unless we can determine they shouldn't
//  - if we got a /test all or an /ok-to-test, we want to consider any job
//    that doesn't explicitly require a human trigger comment; jobs will
//    default to not run unless we can determine that they should
// If a comment that we get matches more than one of the above patterns, we
// consider the set of matching presubmits the union of the results from the
// matching cases.
func FilterPresubmits(honorOkToTest bool, gitHubClient GitHubClient, body string, pr *github.PullRequest, presubmits []config.Presubmit, logger *logrus.Entry) ([]config.Presubmit, []config.Presubmit, error) {
	org, repo, sha := pr.Base.Repo.Owner.Login, pr.Base.Repo.Name, pr.Head.SHA
	filter, err := presubmitFilter(honorOkToTest, gitHubClient, body, org, repo, sha, logger)
	if err != nil {
		return nil, nil, err
	}

	return filterPresubmits(filter, gitHubClient, pr, presubmits, logger)
}

type changesGetter interface {
	GetPullRequestChanges(org, repo string, number int) ([]github.PullRequestChange, error)
}

// filterPresubmits determines which presubmits should run and which should be skipped
// by evaluating the user-provided filter.
func filterPresubmits(filter filter, gitHubClient changesGetter, pr *github.PullRequest, presubmits []config.Presubmit, logger *logrus.Entry) ([]config.Presubmit, []config.Presubmit, error) {
	org, repo, number, branch := pr.Base.Repo.Owner.Login, pr.Base.Repo.Name, pr.Number, pr.Base.Ref
	changes := config.NewGitHubDeferredChangedFilesProvider(gitHubClient, org, repo, number)
	var toTrigger []config.Presubmit
	var toSkipSuperset []config.Presubmit
	for _, presubmit := range presubmits {
		matches, forced, defaults := filter(presubmit)
		if !matches {
			continue
		}
		shouldRun, err := presubmit.ShouldRun(branch, changes, forced, defaults)
		if err != nil {
			return nil, nil, err
		}
		if shouldRun {
			toTrigger = append(toTrigger, presubmit)
		} else {
			toSkipSuperset = append(toSkipSuperset, presubmit)
		}
	}
	toSkip := determineSkippedPresubmits(toTrigger, toSkipSuperset, logger)
	logger.WithFields(logrus.Fields{"to-trigger": toTrigger, "to-skip": toSkip}).Debugf("Filtered %d jobs, found %d to trigger and %d to skip.", len(presubmits), len(toTrigger), len(toSkip))
	return toTrigger, toSkip, nil
}

// determineSkippedPresubmits identifies the largest set of contexts we can actually
// post skipped contexts for, given a set of presubmits we're triggering. We don't
// want to skip a job that posts a context that will be written to by a job we just
// identified for triggering or the skipped context will override the triggered one
func determineSkippedPresubmits(toTrigger, toSkipSuperset []config.Presubmit, logger *logrus.Entry) []config.Presubmit {
	triggeredContexts := sets.NewString()
	for _, presubmit := range toTrigger {
		triggeredContexts.Insert(presubmit.Context)
	}
	var toSkip []config.Presubmit
	for _, presubmit := range toSkipSuperset {
		if triggeredContexts.Has(presubmit.Context) {
			logger.WithFields(logrus.Fields{"context": presubmit.Context, "job": presubmit.Name}).Debug("Not skipping job as context will be created by a triggered job.")
			continue
		}
		toSkip = append(toSkip, presubmit)
	}
	return toSkip
}

type statusGetter interface {
	GetCombinedStatus(org, repo, ref string) (*github.CombinedStatus, error)
}

func presubmitFilter(honorOkToTest bool, statusGetter statusGetter, body, org, repo, sha string, logger *logrus.Entry) (filter, error) {
	// the filters determine if we should check whether a job should run, whether
	// it should run regardless of whether its triggering conditions match, and
	// what the default behavior should be for that check. Multiple filters
	// can match a single presubmit, so it is important to order them correctly
	// as they have precedence -- filters that override the false default should
	// match before others. We order filters by amount of specificity.
	var filters []filter
	filters = append(filters, commandFilter(body))
	if retestRe.MatchString(body) {
		logger.Debug("Using retest filter.")
		combinedStatus, err := statusGetter.GetCombinedStatus(org, repo, sha)
		if err != nil {
			return nil, err
		}
		allContexts := sets.NewString()
		failedContexts := sets.NewString()
		for _, status := range combinedStatus.Statuses {
			allContexts.Insert(status.Context)
			if status.State == github.StatusError || status.State == github.StatusFailure {
				failedContexts.Insert(status.Context)
			}
		}

		filters = append(filters, retestFilter(failedContexts, allContexts))
	}
	if (honorOkToTest && okToTestRe.MatchString(body)) || testAllRe.MatchString(body) {
		logger.Debug("Using test-all filter.")
		filters = append(filters, testAllFilter())
	}
	return aggregateFilter(filters), nil
}

// filter digests a presubmit config to determine if:
//  - we can be certain that the presubmit should run
//  - we know that the presubmit is forced to run
//  - what the default behavior should be if the presubmit
//    runs conditionally and does not match trigger conditions
type filter func(p config.Presubmit) (shouldRun bool, forcedToRun bool, defaultBehavior bool)

// commandFilter builds a filter for `/test foo`
func commandFilter(body string) filter {
	return func(p config.Presubmit) (bool, bool, bool) {
		return p.TriggerMatches(body), p.TriggerMatches(body), true
	}
}

// retestFilter builds a filter for `/retest`
func retestFilter(failedContexts, allContexts sets.String) filter {
	return func(p config.Presubmit) (bool, bool, bool) {
		return failedContexts.Has(p.Context) || (!p.NeedsExplicitTrigger() && !allContexts.Has(p.Context)), false, true
	}
}

// testAllFilter builds a filter for the automatic behavior of `/test all`.
// Jobs that explicitly match `/test all` in their trigger regex will be
// handled by a commandFilter for the comment in question.
func testAllFilter() filter {
	return func(p config.Presubmit) (bool, bool, bool) {
		return !p.NeedsExplicitTrigger(), false, false
	}
}

// aggregateFilter builds a filter that evaluates the child filters in order
// and returns the first match
func aggregateFilter(filters []filter) filter {
	return func(presubmit config.Presubmit) (bool, bool, bool) {
		for _, filter := range filters {
			if shouldRun, forced, defaults := filter(presubmit); shouldRun {
				return shouldRun, forced, defaults
			}
		}
		return false, false, false
	}
}
