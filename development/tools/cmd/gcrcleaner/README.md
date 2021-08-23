# GCR cleaner

## Overview

This command finds and removes old GCR images created by jobs in the Google Cloud Platform (GCP) project.

There are three conditions used to find images for removal:
- The repository name pattern is not on the ignored list.
- The `creationTimestamp` value of the images, that is used to find addresses, exists at least for a preconfigured number of hours.

GCR images that meet these conditions are subject to removal.

## Usage

For safety reasons, the dry-run mode is the default one.
To run it, use:
```bash
env GOOGLE_APPLICATION_CREDENTIALS={path to service account file} go run main.go \
    --repository={gcloud repository url}
```

To turn the dry-run mode off, use:
```bash
env GOOGLE_APPLICATION_CREDENTIALS={path to service account file} go run main.go \
    --repository={gcloud repository url} \
    --dry-run=false
```

## Flags

See the list of available flags:

| Name                      | Required | Description                                                                                          |
| :------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **--repository**          |   YES    | GCR repository name.
| **--dry-run**             |    No    | The boolean value that controls the dry-run mode. It defaults to `true`.
| **--age-in-hours**         |    No    | The integer value for the number of hours. It only matches images older than `now()-ageInHours`. It defaults to `24`.
| **--gcr-exclude-name-regex**       |    YES    | The string value with a valid Golang regexp. It is used to exclude matched repositories by their name.

### Environment variables

See the list of available environment variables:

| Name                                  | Required | Description                                                                                          |
| :------------------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **GOOGLE_APPLICATION_CREDENTIALS**    |    YES   | The path to the service account file. The service account requires at least `browser` and `roles/storage.legacyBucketOwner` Google IAM roles. |
