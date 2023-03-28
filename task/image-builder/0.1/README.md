# Image Builder

The image-builder Task builds an OCI image from a Dockerfile.
Image is signed with signify and pushed to kyma repositories.
The Task use k8s prow image-builder tool to build the image.
Documentation of image builder image used by this task : https://github.com/kyma-project/test-infra/tree/main/development/image-builder

## Compatibility

- **Tekton** v0.36.0 and above

## Install

```shell
kubectl apply -f https://raw.githubusercontent.com/kyma-project/test-infra/main/task/image-builder/0.1/image-builder.yaml
```

## Workspaces

- **`repo`**: The workspace stores sources to build an image from. Usually this should be a shared workspace with other
  tasks. _(REQUIRED)_
- **`config`**: The workspace stores the image-builder config file. It's mounted from configmap. _(REQUIRED)_
- **`signify-secret`**: The workspace stores signify credentials. It's mounted from secret. _(REQUIRED)_
- **`image-registry-credentials`**: The workspace storing image registry credentials. It's mounted from secret. _(
  REQUIRED)_

## Parameters

- **`name`**: Name of the image to be built _(REQUIRED)_
- **`config`**: Path to image-builder config file. _(OPTIONAL, default: "/config/kaniko-build-config.yaml")_
- **`context`**: Path to build context. _(OPTIONAL, default: ".")_
- **`dockerfile`**: Path to Dockerfile file relative to context. _(REQUIRED)_
- **`JOB_TYPE`**: Type of prowjob. Variable set by prow according. _(REQUIRED)_
- **`PULL_NUMBER`**: Pull request number. Variable set by prow. _(REQUIRED)_
- **`PULL_BASE_SHA`**: Git SHA of the base branch. Variable set by prow. _(REQUIRED)_
- **`CI`**: Set to true when the current environment is a CI environment. Variable set by prow. _(REQUIRED)_
- **`REPO_OWNER`**: GitHub org that triggered the job. Variable set by prow. _(REQUIRED)_
- **`REPO_NAME`**: GitHub repo that triggered the job. Variable set by prow. _(REQUIRED)_

## Platforms

The Pipeline can be run on `linux/amd64` platform.

## Usage

See the following samples for usage:

- **[`prowjob-building-image.yaml`](samples/prowjob-building-image.yaml)**: A presubmit prowjob building,signing and
  pushing an image.

## Contributing

We ❤ contributions.

This task is maintained at [kyma-project/test-infra](https://github.com/kyma-project/test-infra). Issues, pull requests
and other contributions can be made there.

To learn more, read the [CONTRIBUTING][contributing] document.

[contributing]: https://github.com/kyma-project/test-infra/blob/main/CONTRIBUTING.md
