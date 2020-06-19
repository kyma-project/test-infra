# image-syncer

## Overview

image-syncer is used to **safely** copy container images from one registry to another. 
The main use case is to preserve images from third party registries in our own registry that we can rely on.

Syncing process steps:
1. Pull image from source.
2. Check if image exists in target.
3. If image does not exist: re-tag image and push to target registry.  
If image exists: compare IDs. If they are different, synchronization is interrupted. It means that something is wrong and the source image was changed (this should not happen).

These steps guarantee that images in our registry are immutable.

## Usage

To run image-syncer, use:
```bash
go run main.go \ 
    --images-file={PATH_TO_A_YAML_FILE_CONTAINING_SYNC_DEFINITION} \
    --target-key-file={PATH_TO_A_JSON_KEY_FILE} \
    --dry-run=true
```

### Definition file

image-syncer takes as an input parameter a file having the following structure: 

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
| **SYNCER_IMAGES_FILE**         |    Yes   | Path to the YAML file with the sync definition, provided as a string.       |
| **SYNCER_TARGET_KEY_FILE**     |    Yes   | Path to the JSON key file, provided as a string.                        |
| **SYNCER_DRY_RUN**             |    No    | Value controlling the `dry run` mode, provided as a boolean.                     |
