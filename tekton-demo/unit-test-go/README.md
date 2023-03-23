# Unit Tests for Go Code

A unit-test-go Task is used to run unit tests for go code.
Task is tailored to be triggered by prow as a prowjob.
It uses a golangci-lint tool to run the tests.
The golangci tool config is stored in the .golangci.yaml file in a repository hosting tested code.

## Compatibility

- **Tekton** v0.36.0 and above

## Install

```shell
kubectl apply -f https://raw.githubusercontent.com/kyma-project/test-infra/main/task/unit-tests-go/0.1/unit-tests-go.yaml
```

## Workspaces

- **`repo`**: The workspace where sources to test are stored. Usually this it should be a shared workspace with other
  tasks. _(REQUIRED)_

## Parameters

- **`path-to-test`**: The path to go code to test. _(OPTIONAL, default: "./...")_
- **`REPO_OWNER`**: GitHub org that triggered the prowjob. Variable set by prow. Prow will not set this variable if the
  prowjob type is periodic. In that case an empty string default value will be used. _(OPTIONAL, default: "")_
- **`REPO_NAME`**: GitHub repo that triggered the prowjob. Variable set by prow. Prow will not set this variable if the
  prowjob type is periodic. In that case an empty string default value will be used. _(OPTIONAL, default: "")_

## Platforms

The Pipeline can be run on `linux/amd64` platform.

## Usage

See the following samples for usage:

- **[`cache-image.yaml`](samples/cache-image.yaml)**: A PipelineRun configured to cache build artifacts in an image.
- **[`cache-volume.yaml`](samples/cache-volume.yaml)**: A PipelineRun configured to cache build artifacts in a volume.
- **[`env-vars.yaml`](samples/env-vars.yaml)**: A PipelineRun configured to provide _build-time_ environment variables.
- **[`run-image.yaml`](samples/run-image.yaml)**: A PipelineRun configured to specify an explicit run image.

## Contributing

We ❤ contributions.

This task is maintained at [kyma-project/test-infra](https://github.com/kyma-project/test-infra). Issues, pull requests
and other contributions can be made there.

To learn more, read the [CONTRIBUTING][contributing] document.

[contributing]: https://github.com/kyma-project/test-infra/blob/main/CONTRIBUTING.md
