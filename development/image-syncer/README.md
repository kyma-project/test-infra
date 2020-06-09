# Image Syncer

## Overview

Image Syncer is used to **safely** copy container images from one registry to another. 
The main use-case is to preserve images from third party registries in our own registry that we can relay on.

Syncing process steps:
- pull image from source
- check if image exists in target
- if it does not exist - re-tag image and push to target registry
- if it exists compare IDs - if they are different synchronisation is interrupted, 
it means that something is wrong - source image was changed (that should not happen)

These steps guarantee that images in our registry are immutable.

## Usage

To run it, use:
```bash
go run main.go \ 
    --images-file={path to a yaml file containing sync definition} \
    --target-key-file={path to a json key file} \
    --dry-run=true
```

***Defintion file***

Image syncer as an input parameter takes a file containing following structure: 

```yaml
targetRepoPrefix:  "eu.gcr.io/kyma-project/external/"
images:
- source: "bitnami/keycloak-gatekeeper:9.0.3"
- source: "busybox:1.31.1"

```

### Flags

```
Usage:
  image-syncer [flags]

Flags:
  -d, --dry-run                  dry run mode [SYNCER_DRY_RUN] (default true)
  -h, --help                     help for githubstats
  -i, --images-file string       yaml file containing list of images [SYNCER_IMAGES_FILE]
  -t, --target-key-file string   JSON key file used for authorization to target repo [SYNCER_TARGET_KEY_FILE]

exit status 1
```


### Environment variables

All flags can also be set using these environment variables:

| Name                           | Required | Description                                                           |
| :----------------------------- | :------: | :-------------------------------------------------------------------- |
| **SYNCER_IMAGES_FILE**         |    Yes   | The string value with a path to yaml file with sync definition.       |
| **SYNCER_TARGET_KEY_FILE**     |    Yes   | The string value with a path to json key file.                        |
| **SYNCER_DRY_RUN**             |    No    | The boolean value controlling the `dry run` mode.                     |
