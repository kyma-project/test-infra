# Unit Tests for Go Code

A go-unit-tests Task is used to run unit tests for Go code.
The task is tailored to be triggered by Prow as a ProwJob.
It uses a built-in `go test` command to run unit tests.

## Compatibility

- Tekton v0.36.0 and above

## Install

```shell
kubectl apply -f https://raw.githubusercontent.com/kyma-project/test-infra/main/task/go-unit-tests/0.1/go-unit-tests.yaml
```

## Workspaces

- **repo**: The workspace where tested sources are stored. Usually, this should be a workspace shared with other
  tasks. _(REQUIRED)_

## Parameters

- **path-to-test**: The path to tested Go code. _(OPTIONAL, default: "./...")_
- **REPO_OWNER**: GitHub org that triggers the ProwJob. A variable set by Prow. Prow does not set this variable if the
  ProwJob type is periodic. In that case, an empty string default value is used. _(OPTIONAL, default: "")_
- **REPO_NAME**: GitHub repo that triggers the ProwJob. A variable set by Prow. Prow does not set this variable if the
  ProwJob type is periodic. In that case, an empty string default value is used. _(OPTIONAL, default: "")_

## Platforms

You can run the Task on the `linux/amd64` platform.

## Usage

See the following samples for usage:

- [`prowjob-testing-go-code.yaml`](samples/prowjob-testing-go-code.yaml): A presubmit ProwJob configured to run Golang unit tests.

## Contributing

We ❤ contributions.

This task is maintained in the [`Test Infra`](https://github.com/kyma-project/test-infra) repository. Issues, pull requests
and other contributions can be made there.

To learn more, read the [CONTRIBUTING][contributing] document.

[contributing]: https://github.com/kyma-project/test-infra/blob/main/CONTRIBUTING.md
