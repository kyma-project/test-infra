# image-syncer

## Overview

The image-syncer is used to copy container images between two registries.
It copies images **only when they are not present** in the target repo.
The tool guarantees that **tags are immutable** in the target repo.
That means that if the image tag is already present in the target registry, it will not be overwritten by the new image.
The tool can be used in a GitHub workflow to synchronize images defined in a YAML file maintained in the repository.

## User Guide

The developers can use image-syncer to synchronize images defined in the YAML file maintained in their repository.
Each team defines the list of images they want to synchronize.

Follow the steps below to use the image-syncer for your repository:

1. Create a workflow that calls
   the [image-syncer reusable workflow](https://github.com/kyma-project/test-infra/blob/main/.github/workflows/image-syncer.yml).
   See the example [workflow](#example-workflow) below.
2. Create a PR with the YAML file with the list of images you want to synchronize.
   See the example [images list file](#example-images-list-file) below.
3. Merge the PR to synchronize the images.

### Workflow using image-syncer

> [!IMPORTANT]
> The image-syncer running for pull request will always run in dry-run mode.
> The workflow calling image-syncer reusable workflow on pull request must use the pull_request_target event as a trigger.

See
supported [inputs](https://github.com/kyma-project/test-infra/blob/4df11c5384a5c7ac3ce76b726e17dee6aba07f79/.github/workflows/image-syncer.yml#L5)
by the image-syncer reusable workflow.

The image-syncer supports following github events:

- pull_request_target
- push

The workflow must call the image-syncer reusable workflow from the main branch of the test-infra repository.
The value of uses must be `kyma-project/test-infra/.github/workflows/image-syncer.yml@main`.

#### Example Workflow

```yaml
name: pull-sync-external-images

on:
  pull_request_target:
    branches:
      - main
    types: [ opened, edited, synchronize, reopened, ready_for_review ]
    paths:
      - "external-images.yaml"

permissions:
  id-token: write # This is required for requesting the JWT token
  contents: read # This is required for actions/checkout

jobs:
  sync-external-images:
    uses: kyma-project/test-infra/.github/workflows/image-syncer.yml@main
    with:
      debug: true
```

### Example Images List File

> [!IMPORTANT]
> The image-syncer expects the file external-images-yaml with the list of images in root directory of the repository.

As an input parameter, image-syncer takes a file having the following structure:

- source: the source image to be copied. It can be an image name with a tag or a digest.
- tag: the tag of the image in the target repository. This is required when the source uses digest to identify the image.
- amd64Only: a boolean value that indicates if it's allowed to synchronize the image manifest instead of image index.
  This is required when the source image is not a multi-platform image.

```yaml
images:
- source: "bitnami/keycloak-gatekeeper:9.0.3"
- source: "bitnami/postgres-exporter:0.8.0-debian-10-r28"
- source: "busybox@sha256:31a54a0cf86d7354788a8265f60ae6acb4b348a67efbcf7c1007dd3cf7af05ab"
  tag: "1.32.0-v1"
- source: "bitnami/postgres-exporter:0.11.1-debian-11-r69"
  amd64Only: true
- source: "postgres@sha256:9d7ec48fe46e8bbce55deafff58080e49d161a3ed92e67f645014bb50dc599fd"
  tag: "v20230508-11.19-alpine3.17"
  amd64Only: true
```

## Developer Guide

### Reusable Workflow

### Syncing Process

Syncing process steps:

1. Pull an image from source.
2. Check if the image name contains the multi-platform SHA256 digest. To get this digest,
   run `docker run mplatform/manifest-tool inspect {your-image}:{image-version}` and copy the first digest this command returns. For the
   non-multi-platform images supporting only amd64 architecture, use their amd64 digest and add the `amd64only: true` parameter.
3. If the image name contains digest, re-tag the target image with the tag instead of SHA256 digest.
4. Check if the image exists in target.
5. If the image does not exist, re-tag the image and push to the target registry.  
   If the image exists, compare the IDs. If they are different, synchronization is interrupted. It means that something is wrong and the
   source image was changed (this should not happen).
6. Push the signature to the target repository.

These steps guarantee that images in your registry are immutable and verified.

### Image Syncer Container Image

### Flags

```
Usage:
  image-syncer [flags]

Flags:
      --debug                         Enables the debug mode [SYNCER_DEBUG]
      --dry-run                       Enables the dry-run mode [SYNCER_DRY_RUN]
  -h, --help                          help for image-syncer
  -i, --images-file string            Specifies the path to the YAML file that contains list of images [SYNCER_IMAGES_FILE]
  -t, --target-repo-auth-key string   Specifies the JSON key file used for authorization to the target repository [SYNCER_TARGET_REPO_AUTH_KEY]
```


### Environment Variables

All the flags can also be set using these environment variables:

| Name                           | Required | Description                                                           |
| :----------------------------- | :------: | :-------------------------------------------------------------------- |
| **SYNCER_IMAGES_FILE**         |    Yes   | Path to the YAML file with the sync definition, provided as a string.|
| **SYNCER_TARGET_REPO_AUTH_KEY**|    Yes   | Path to the JSON key file, provided as a string.|
| **SYNCER_DRY_RUN**             |    No    | Value controlling the `dry run` mode, provided as a boolean.|
|**SYNCER_DEBUG**| No | Variable that enables the `debug` mode.|