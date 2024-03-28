# GCR Cleaner

## Overview

This command finds and removes old GCR images created by Jobs in the Google Cloud project.

There are two conditions used to find images for removal:
- The repository name pattern is not on the ignored list.
- The `creationTimestamp` value of the images, which is used to find addresses, exists at least for a preconfigured number of hours.

GCR images that meet all these conditions are subject to removal.

## Usage

For safety reasons, the dry-run mode is the default one.  
To run it, use:
```bash
env GOOGLE_APPLICATION_CREDENTIALS={PATH_TO_SERVICE_ACCOUNT_FILE} go run main.go \
    --repository={GCLOUD_REPOSITORY_URL}
```

To turn the dry-run mode off, use:
```bash
env GOOGLE_APPLICATION_CREDENTIALS={PATH_TO_SERVICE_ACCOUNT_FILE} go run main.go \
    --repository={GCLOUD_REPOSITORY_URL} \
    --dry-run=false
```

### Flags

See the list of available flags:

| Name                      | Required | Description                                                                                          |
| :------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **--repository**          |   Yes    | The GCR repository name.
| **--dry-run**             |    No    | The boolean value that controls the dry-run mode. Defaults to `true`.
| **--age-in-hours**         |    No    | The integer value for the number of hours. It only matches images older than `now()-ageInHours`. Defaults to `24`.
| **--gcr-exclude-name-regex**       |    Yes    | The string value with a valid Go regexp. Used to exclude matched repositories by their name.

### Environment Variables

See the list of available environment variables:

| Name                                  | Required | Description                                                                                          |
| :------------------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **GOOGLE_APPLICATION_CREDENTIALS**    |    Yes   | The path to the service account file. The service account requires at least the `browser` and `roles/storage.legacyBucketOwner` Google IAM roles. |
