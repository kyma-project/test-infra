# Unit Tests for Go Code

The golangci-lint Task runs a Golang linter.
The task is tailored to be triggered by Prow as a ProwJob.
It uses a golangci-lint tool.
The GolangCI tool config is stored in the `.golangci.yaml` file in a repository hosting tested code.
The task does not accept additional parameters to configure the linter. By default, it runs linter against all modules
found under task working directory.

## Compatibility

- Tekton v0.36.0 and above

## Install

```shell
kubectl apply -f https://raw.githubusercontent.com/kyma-project/test-infra/main/task/golangci-lint/0.1/golangci-lint.yaml
```

## Workspaces

- **repo**: The workspace where tested sources are stored. Usually, this should be a workspace shared  with other
  tasks. _(REQUIRED)_

## Parameters

- **`REPO_OWNER`**: GitHub org that triggers the ProwJob. A variable set by Prow. Prow does not set this variable if the
  ProwJob type is periodic. In that case, an empty string default value is used.  _(OPTIONAL, default: "")_
- **`REPO_NAME`**: GitHub repo that triggers the ProwJob. A variable set by Prow. Prow does not set this variable if the
  ProwJob type is periodic. In that case, an empty string default value is used. _(OPTIONAL, default: "")_

## Platforms

You can run the Task on the `linux/amd64` platform.

## Usage

See the following samples for usage:

- [`prowjob-linting-go-code.yaml`](samples/prowjob-linting-go-code.yaml): A presubmit ProwJob configured to run
  a Golang linter.

## Contributing

We ❤ contributions.

This task is maintained in the [`Test Infra`](https://github.com/kyma-project/test-infra) repository. Issues, pull requests
and other contributions can be made there.

To learn more, read the [CONTRIBUTING][contributing] document.

[contributing]: https://github.com/kyma-project/test-infra/blob/main/CONTRIBUTING.md
