# IP cleaner

## Overview

This command finds and removes orphaned IP addresses created by jobs in the Google Cloud project.

There are three conditions used to find addresses for removal:
- The address name pattern is not on the ignored list.
- The **users** field of the address shows `0`, which means that the disk is unused.
- The `creationTimestamp` value of the address, that is used to find addresses, exists at least for a preconfigured number of hours.

IP addresses that meet all these conditions are subject to removal.

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
| **--project**             |   YES    | Google Cloud project name.
| **--dry-run**             |    No    | The boolean value that controls the dry-run mode. It defaults to `true`.
| **--age-in-hours**         |    No    | The integer value for the number of hours. It only matches disks older than `now()-ageInHours`. It defaults to `2`.
| **--ip--exclude-name-regex**       |    No    | The string value with a valid Golang regexp. It is used to exclude matched addresses by their name. It defaults to `^nightly|weekly|nat-auto-ip`.

### Environment Variables

See the list of available environment variables:

| Name                                  | Required | Description                                                                                          |
| :------------------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **GOOGLE_APPLICATION_CREDENTIALS**    |    YES   | The path to the service account file. The service account requires at least `container.clusters.list` and `container.clusters.delete` Google IAM permissions. |

