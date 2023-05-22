# Hide GitHub Comments

The hide-github-comments Task hides comments which belong to an old commit.
The tool detects old comments by reading github-comments metadata in GitHub comments.

## Compatibility

- **Tekton** v0.36.0 and above

## Install

```shell
kubectl apply -f https://raw.githubusercontent.com/kyma-project/test-infra/main/task/hide-github-comments/0.1/hide-github-comments.yaml
```

## Parameters

- **github-token-secret**: Name of the secret holding the Kyma bot GitHub token. _(OPTIONAL, default: 'kyma-bot-github-token')_
- **PULL_NUMBER**: Pull request number. A variable set by Prow. _(REQUIRED)_
- **PULL_PULL_SHA**: Git SHA of the pull request head branch. A variable set by Prow.  _(REQUIRED)_
- **REPO_OWNER**: The GitHub organization that triggers the job. A variable set by Prow.  _(REQUIRED)_
- **REPO_NAME**: The GitHub repository that triggers the job. A variable set by Prow.  _(REQUIRED)_

## Platforms

You can run the Task on the `linux/amd64` platform.

## Usage

See the following samples for usage:

- [`prowjob-building-image.yaml`](samples/sample_prowjob_pipeline.yaml): A presubmit ProwJob that builds, signs, and pushes an image.

## Contributing

We ❤ contributions.

This task is maintained in the [`Test Infra`](https://github.com/kyma-project/test-infra) repository. Issues, pull requests and other contributions can be made there.

To learn more, read the [CONTRIBUTING][contributing] document.

[contributing]: https://github.com/kyma-project/test-infra/blob/main/CONTRIBUTING.md
