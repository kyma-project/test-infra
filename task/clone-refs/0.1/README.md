# Unit Tests for Go Code

The clonerefs Task clone repositories references defined by prowjob refs and extra-refs.
Prowjob refs and extra-refs are read from the environment variable JOB_SPEC.
JOB_SPEC is a JSON-encoded prowjob specification and is set by prow.
The Task will clone the repositories in to workspace repo.
The Task use k8s prow clonerefs tool to clone the repositories.

## Compatibility

- **Tekton** v0.36.0 and above

## Install

```shell
kubectl apply -f https://raw.githubusercontent.com/kyma-project/test-infra/main/task/clone-refs/0.1/clone-refs.yaml
```

## Workspaces

- **`repo`**: The workspace where the git repos will be cloned. Usually this it should be a shared workspace with other
  tasks. _(REQUIRED)_
- **`logs`**: "The workspace where the clone logs will be written." _(OPTIONAL, default: "/logs")_
- **`tmp`**: The workspace for temporary files. _(OPTIONAL, default: "/tmp")_

## Parameters

- **`JOB_SPEC`**: JSON-encoded job specification. Variable set by prow. _(REQUIRED)_
- **`LOG`**: The path to the clone logs file. _(OPTIONAL, default: "/logs/clone.json")_
  prowjob type is periodic. In that case an empty string default value will be used. _(OPTIONAL, default: "")_
- **`SRC_ROOT`**: The root path where to clone repositories. _(OPTIONAL, default: "/home/prow/go")_

## Platforms

The Pipeline can be run on `linux/amd64` platform.

## Usage

See the following samples for usage:

- **[`prowjob-cloning-repositories.yaml`](samples/prowjob-cloning-repositories.yaml)**: A presubmit prowjob cloning
  repositories from github.

## Contributing

We ❤ contributions.

This task is maintained at [kyma-project/test-infra](https://github.com/kyma-project/test-infra). Issues, pull requests
and other contributions can be made there.

To learn more, read the [CONTRIBUTING][contributing] document.

[contributing]: https://github.com/kyma-project/test-infra/blob/main/CONTRIBUTING.md
