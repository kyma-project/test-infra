# Unit Tests for Go Code

The golangci-lint Task runs a golang linter.
Task is tailored to be triggered by prow as a prowjob.
It uses a golangci-lint tool.
The golangci tool config is stored in the .golangci.yaml file in a repository hosting tested code.

## Compatibility

- **Tekton** v0.36.0 and above

## Install

```shell
kubectl apply -f https://raw.githubusercontent.com/kyma-project/test-infra/main/task/golangci-lint/0.1/golangci-lint.yaml
```

## Workspaces

- **`repo`**: The workspace where sources to test are stored. Usually this it should be a shared workspace with other
  tasks. _(REQUIRED)_

## Parameters

- **`REPO_OWNER`**: GitHub org that triggered the prowjob. Variable set by prow. Prow will not set this variable if the
  prowjob type is periodic. In that case an empty string default value will be used. _(OPTIONAL, default: "")_
- **`REPO_NAME`**: GitHub repo that triggered the prowjob. Variable set by prow. Prow will not set this variable if the
  prowjob type is periodic. In that case an empty string default value will be used. _(OPTIONAL, default: "")_

## Platforms

The Pipeline can be run on `linux/amd64` platform.

## Usage

See the following samples for usage:

- **[`prowjob-linting-go-code.yaml`](samples/prowjob-linting-go-code.yaml)**: A presubmit prowjob configured to run
  golang linter.

## Contributing

We ❤ contributions.

This task is maintained at [kyma-project/test-infra](https://github.com/kyma-project/test-infra). Issues, pull requests
and other contributions can be made there.

To learn more, read the [CONTRIBUTING][contributing] document.

[contributing]: https://github.com/kyma-project/test-infra/blob/main/CONTRIBUTING.md
