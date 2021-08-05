# IP cleaner

## Overview

This command finds and removes orphaned IP addresses created by the jobs in a Google Cloud Platform (GCP) project.

There are three conditions used to find addresses for removal:
- The address name pattern is not on the ignored list
- The address users count where zero means the disk is unused
- The address `creationTimestamp` value that is used to find addresses existing at least for a preconfigured number of hours

IP addresses that meet these conditions are subject to removal.

## Usage

For safety reasons, the dry-run mode is the default one.
To run it, use:
```bash
env GOOGLE_APPLICATION_CREDENTIALS={path to service account file} go run main.go \
    --project={gcloud project name}
```

To turn the dry-run mode off, use:
```bash
env GOOGLE_APPLICATION_CREDENTIALS={path to service account file} go run main.go \
    --project={gcloud project name} \
    --dry-run=false
```

### Flags

See the list of available flags:

| Name                      | Required | Description                                                                                          |
| :------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **--project**             |   YES    | GCP project name.
| **--dry-run**             |    No    | The Boolean value that controls the dry-run mode. It defaults to `true`.
| **--age-in-hours**         |    No    | The integer value for the number of hours. It only matches disks older than `now()-ageInHours`. It defaults to `2`.
| **--ip-name-regex**       |    No    | The string value with a valid Golang regexp. It is used to exclude matched addresses by their name. It defaults to `^nightly|weekly|nat-auto-ip`.
### Environment variables

See the list of available environment variables:

| Name                                  | Required | Description                                                                                          |
| :------------------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **GOOGLE_APPLICATION_CREDENTIALS**    |    YES   | The path to the service account file. The service account requires at least `container.clusters.list` and `container.clusters.delete` Google IAM permissions. |

