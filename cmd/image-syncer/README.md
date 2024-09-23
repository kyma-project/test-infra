# image-syncer

Contents:

- [Overview](#overview)
- [User Guide](#user-guide)
    - [Workflow](#workflow)
        - [Example Workflow](#example-workflow)
    - [Example Images List File](#example-images-list-file)
- [Developer Guide](#developer-guide)
    - [Syncing Process](#syncing-process)
    - [Reusable Workflow](#reusable-workflow)
    - [Image Syncer Container Image](#image-syncer-container-image)
        - [Flags](#flags)
        - [Environment Variables](#environment-variables)
    - [Infrastucture as Code](#infrastucture-as-code)

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

### Workflow

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

The image-syncer tool consists of the following components:

- The image-syncer binary, written in Go. The binary is released as a container image.
- The image-syncer reusable workflow that can be called from the GitHub workflow.
- The image-syncer infrastructure resources are used to access the target registry.

### Image Syncer Binary

The image-syncer binary is responsible for synchronizing images between two registries.
The binary is released as a container image.
The [Dockerfile](https://github.com/kyma-project/test-infra/blob/main/cmd/image-syncer/Dockerfile) for the image-syncer binary.

> [!IMPORTANT]
> The amd64Only field in the images list file allows syncing the image manifest instead of the image index.
> The flag does not prevent syncing other platforms.

#### Flags

The image-syncer binary accepts flags to configure the synchronization process.
Supported [flags](https://github.com/kyma-project/test-infra/blob/1df13d56ad523ce434e33284bb7e392ff897cd1b/cmd/image-syncer/main.go#L274-L282).

#### Environment Variables

All the flags can also be set using environment variables.
The environment variables must be prefixed with `SYNCER` and use uppercase letters.
Dash characters must be replaced with underscores.
For example, the `--dry-run` flag can be set using the `SYNCER_DRY_RUN` environment variable.

### Reusable Workflow

The image-syncer reusable workflow provides an execution environment for the image-syncer binary.
The workflow is called from the GitHub workflow defined in other repositories.
It authenticates in Google Cloud and provides the image-syncer binary with the required credentials.

### Infrastucture as Code

The image-syncer needs certain infrastructure resources to access the target registry.
The resources are defined in
the [infrastructure-as-code](https://github.com/kyma-project/test-infra/tree/main/configs/terraform/environments/prod) configuration.

- The image-syncer [resources](https://github.com/kyma-project/test-infra/blob/main/configs/terraform/environments/prod/image-syncer.tf).
- The
  image-syncer [varaibles](https://github.com/kyma-project/test-infra/blob/main/configs/terraform/environments/prod/image-syncer-variables.tf).