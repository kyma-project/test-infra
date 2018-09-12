## Overview

Investigation results for the plugins and additional tooling used by Prow. This investigation was done on 12 September 2018.

### Prow Plugins[1]

**approve**

The approve plugin implements a pull request approval process that manages the 'approved' label and an approval notification comment.

**assign**

The assign plugin assigns or requests reviews from users.

**blockade**

The blockade plugin blocks pull requests from merging if they touch specific files.

**blunderbuss**

The blunderbuss plugin automatically requests reviews from reviewers when a new PR is created.

**cat**

The cat plugin adds a cat image to an issue in response to the `/meow` command.

**cherry-pick-unapproved**

Label PRs against a release branch which do not have the `cherry-pick-approved` label with the `do-not-merge/cherry-pick-not-approved` label.

**cla**

The cla plugin manages the application and removal of the 'cncf-cla' prefixed labels on pull requests as a reaction to the cla/linuxfoundation github status context.

**config-updater**

The config-updater plugin automatically redeploys configuration and plugin configuration files when they change.

**docs-no-retest**

The docs-no-retest plugin applies the 'retest-not-required-docs-only' label to pull requests that only touch documentation type files and thus do not need to be retested against the latest master commit before merging.

**dog**

The dog plugin adds a dog image to an issue in response to the `/woof` command.

**golint**

The golint plugin runs golint on changes made to *.

**heart**

The heart plugin celebrates certain Github actions with the reaction emojis.

**help**

The help plugin provides commands that add or remove the 'help wanted' and the 'good first issue' labels from issues.

**hold**

The hold plugin allows anyone to add or remove the 'do-not-merge/hold' label from a pull request in order to temporarily prevent the PR from merging without withholding approval.

**label**

The label plugin provides commands that add or remove certain types of labels.

**lgtm**

The lgtm plugin manages the application and removal of the 'lgtm' (Looks Good To Me) label which is typically used to gate merging.

**lifecycle**

Close, reopen, flag and/or unflag an issue or PR as frozen/stale/rotten.

**milestone**

The milestone plugin allows members of a configurable GitHub team to set the milestone on an issue or pull request.

**milestonestatus**

The milestonestatus plugin allows members of the milestone maintainers Github team to specify the 'status/*' label that should apply to a pull request.

**needs-rebase**
The needs-rebase plugin manages the 'needs-rebase' label by removing it from Pull Requests that are mergeable and adding it to those which are not.

**override**

The override plugin allows repo admins to force a github status context to pass.

**owners-label**

The owners-label plugin automatically adds labels to PRs based on the files they touch.

**release-note**

The releasenote plugin implements a release note process that uses a markdown 'releasenote' code block to associate a release note with a pull request.

**require-matching-label**

The require-matching-label plugin is a configurable plugin that applies a label to issues and/or PRs that do not have any labels matching a regular expression.

**require-sig**

When a new issue is opened the require-sig plugin adds the "needs-sig" label and leaves a comment requesting that a SIG (Special Interest Group) label be added to the issue.

**shrug**

¯\_(ツ)_/¯

**sigmention**

The sigmention plugin responds to SIG (Special Interest Group) Github team mentions like '@kubernetes/sig-testing-bugs'.

**size**

The size plugin manages the 'size/*' labels, maintaining the appropriate label on each pull request as it is updated.

**skip**

The skip plugin allows users to clean up Github stale commit statuses for non-blocking jobs on a PR.

**slackevents**

The slackevents plugin reacts to various Github events by commenting in Slack channels.

**stage**

Label the stage of an issue as alpha/beta/stable.

**trigger**

The trigger plugin starts tests in reaction to commands and pull request events.

**verify-owners**

The verify-owners plugin validates OWNERS files if they are modified in a PR.

**welcome**

The welcome plugin posts a welcoming message when it detects a user's first contribution to a repo.

**wip**

The wip (Work In Progress) plugin applies the 'do-not-merge/work-in-progress' label to pull requests whose title starts with 'WIP' and removes it from pull requests when they remove the title prefix.

**yuks**

The yuks plugin comments with jokes in response to the `/joke` command.

#### Plugins we will use for the first installation
- approve
- assign
- blunderbuss
- config-updater
- docs-no-retest
- help
- hold
- label
- lgtm
- owners-label
- size
- trigger
- verify-owners
- wip

#### Plugins that can be integrated later on demand
- blockade
- cat
- cherry-pick-unapproved
- dog
- lifecycle
- needs-rebase
- override
- owners-label
- release-note
- require-matching-label
- require-sig
- shrug
- sigmention
- skip
- slackevents
- welcome
- yuks

### Additional Tooling used with Prow[2]

**Boskos**

Boskos manages job resources (such as GCP projects) in pools, checking them out for jobs and cleaning them up automatically (with Grafana monitoring).

**ghProxy**

ghProxy is a reverse proxy HTTP cache optimized for use with the GitHub API, to ensure the token usage doesn’t hit API limits (with Grafana monitoring).

**Greenhouse**

Greenhouse allows to use a remote bazel cache to provide faster build and test results for PRs (with Grafana monitoring).

**Gubernator**

Gubernator displays the results and test history for a given PR.

**Kettle**

Kettle transfers data from GCS to a publicly accessible bigquery dataset.

**PR dashboard**

PR dashboard is a workflow-aware dashboard that allows contributors to understand which PRs require attention and why.

**Splice**

Splice allows to test and merge PRs in a batch, ensuring our merge velocity is not limited to our test velocity.

**Testgrid**

Testgrid displays test results for a given job across all runs, summarize test results across groups of jobs.

**Tide**

Tide allows to merge PRs selected via GitHub queries rather than ordered in a queue, allowing for significantly higher merge velocity in tandem with splice.

**Triage**

Triage identifies common failures that happen across all jobs and tests.

#### Additional tooling for the first installation

We will not use any of these tooling for the first installation but we want to integrate Splice and Tide as soon as possible. For that, a thorough analysis should be done regarding the criteria on merging PRs automatically.

## References

- [1]: List and explanations are from [https://prow.k8s.io/plugins](https://prow.k8s.io/plugins).
- [2]: Information has been extracted from [this](https://kubernetes.io/blog/2018/08/29/the-machines-can-do-the-work-a-story-of-kubernetes-testing-ci-and-automating-the-contributor-experience/) blog post.