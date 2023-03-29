# Clone Repositories

The clone-refs Task clones repositories references defined by ProwJob refs and extra-refs.
ProwJob refs and extra-refs are read from the environment variable JOB_SPEC which 
is a JSON-encoded ProwJob specification and is set by Prow.
The Task clones the repositories into a workspace named "repo".
The Task uses the Kubernetes Prow clone-refs tool to clone the repositories.

## Compatibility

- **Tekton** v0.36.0 and above

## Install

```shell
kubectl apply -f https://raw.githubusercontent.com/kyma-project/test-infra/main/task/clone-refs/0.1/clone-refs.yaml
```

## Workspaces

- **repo**: The workspace where the Git repositories are cloned. Usually, this should be a workspace shared with other
  tasks. _(REQUIRED)_
- **logs**: The workspace where the clone logs are written. _(OPTIONAL, default: "/logs")_
- **tmp**: The workspace for temporary files. _(OPTIONAL, default: "/tmp")_

## Parameters

- **JOB_SPEC**: JSON-encoded job specification. A variable set by Prow. _(REQUIRED)_
- **LOG**: The path to the clone logs file. _(OPTIONAL, default: "/logs/clone.json")_
  The ProwJob type is periodic. In that case, an empty string default value is used. _(OPTIONAL, default: "")_
- **SRC_ROOT**: The root path where to clone repositories. _(OPTIONAL, default: "/home/prow/go")_

## Platforms

You can run the Task on `linux/amd64` platform.

## Usage

See the following samples for usage:

- [`prowjob-cloning-repositories.yaml`](samples/prowjob-cloning-repositories.yaml): A presubmit ProwJob cloning
  repositories from GitHub.

## Contributing

We ❤ contributions.

This task is maintained in the  [Test Infra](https://github.com/kyma-project/test-infra) repository. Issues, pull requests
and other contributions can be made there.

To learn more, read the [CONTRIBUTING][contributing] document.

[contributing]: https://github.com/kyma-project/test-infra/blob/main/CONTRIBUTING.md
