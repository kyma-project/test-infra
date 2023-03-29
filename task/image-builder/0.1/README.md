# Image Builder

The image-builder Task builds an OCI image from a Dockerfile.
The image is signed with signify and pushed to Kyma repositories.
The Task uses the Kubernetes Prow image-builder tool to build the image.
See the [documentation](https://github.com/kyma-project/test-infra/tree/main/development/image-builder) of the image-builder image used by this task.

## Compatibility

- **Tekton** v0.36.0 and above

## Install

```shell
kubectl apply -f https://raw.githubusercontent.com/kyma-project/test-infra/main/task/image-builder/0.1/image-builder.yaml
```

## Workspaces

- **repo**: The workspace stores sources for building an image. Usually, this should be a workspace shared with other
  tasks. _(REQUIRED)_
- **config**: The workspace stores the image-builder config file. It's mounted from ConfigMap. _(REQUIRED)_
- **signify-secret**: The workspace stores signify credentials. It's mounted from Secret. _(REQUIRED)_
- **`image-registry-credentials`**: The workspace storing image registry credentials. It's mounted from Secret. _(
  REQUIRED)_

## Parameters

- **name**: The name of the image to be built _(REQUIRED)_
- **config**: The path to the image-builder config file. _(OPTIONAL, default: "/config/kaniko-build-config.yaml")_
- **context**: The path to the build context. _(OPTIONAL, default: ".")_
- **dockerfile**: The path to the Dockerfile file relative to the context. _(REQUIRED)_
- **JOB_TYPE**: The type of ProwJob. A variable set by Prow according to the list of job environment variables. _(REQUIRED)_
- **PULL_NUMBER**: Pull request number. A variable set by Prow. _(REQUIRED)_
- **PULL_BASE_SHA**: Git SHA of the base branch. A variable set by Prow.  _(REQUIRED)_
- **CI**: Set to `true` when the current environment is a CI environment. A variable set by Prow.  _(REQUIRED)_
- **REPO_OWNER**: The GitHub organization that triggers the job. A variable set by Prow.  _(REQUIRED)_
- **REPO_NAME**: The GitHub repository that triggers the job. A variable set by Prow.  _(REQUIRED)_

## Platforms

The Pipeline can be run on `linux/amd64` platform.

## Usage

See the following samples for usage:

- [`prowjob-building-image.yaml`](samples/prowjob-building-image.yaml): A presubmit ProwJob that builds,signs and
  pushes an image.

## Contributing

We ❤ contributions.

This task is maintained in the [Test Infra](https://github.com/kyma-project/test-infra) repository. Issues, pull requests
and other contributions can be made there.

To learn more, read the [CONTRIBUTING][contributing] document.

[contributing]: https://github.com/kyma-project/test-infra/blob/main/CONTRIBUTING.md
