# image-syncer

## Overview

image-syncer is used to copy container images from one registry to another.
The main use case is to preserve images from third party registries in our own registry that we can rely on.

It copies images **only when they are not present** in the target repo. That guarantees that **tags are immutable**.

Syncing process steps:
1. Pull image from source.
2. Check if image name contains SHA256 digest.
3. If image name contains digest, re-tag target image with the tag instead of SHA256 digest.
4. Check if image exists in target.
5. If image does not exist, re-tag image and push to target registry.  
If image exists: compare IDs. If they are different, synchronization is interrupted. It means that something is wrong and the source image was changed (this should not happen).
6. Push the signature to the target repository.

These steps guarantee that images in our registry are immutable and verified.

## Usage

To run image-syncer, use:
```bash
go run main.go \ 
    --images-file={PATH_TO_A_YAML_FILE_CONTAINING_SYNC_DEFINITION} \
    --target-repo-auth-key={PATH_TO_A_JSON_KEY_FILE} \
    --dry-run=true
```
### Definition file

image-syncer takes as an input parameter a file having the following structure: 

```yaml
targetRepoPrefix:  "eu.gcr.io/kyma-project/external/"
images:
- source: "bitnami/keycloak-gatekeeper:9.0.3"
- source: "bitnami/postgres-exporter:0.8.0-debian-10-r28"
- source: "busybox@sha256:31a54a0cf86d7354788a8265f60ae6acb4b348a67efbcf7c1007dd3cf7af05ab"
  tag: "1.32.0-v1"
```

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


### Environment variables

All flags can also be set using these environment variables:

| Name                           | Required | Description                                                           |
| :----------------------------- | :------: | :-------------------------------------------------------------------- |
| **SYNCER_IMAGES_FILE**         |    Yes   | Path to the YAML file with the sync definition, provided as a string.|
| **SYNCER_TARGET_REPO_AUTH_KEY**|    Yes   | Path to the JSON key file, provided as a string.|
| **SYNCER_DRY_RUN**             |    No    | Value controlling the `dry run` mode, provided as a boolean.|
|**SYNCER_DEBUG**|No|Variable that enables the debug mode.|
