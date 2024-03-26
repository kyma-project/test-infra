# Disks Garbage Collector

## Overview

This command finds and removes orphaned disks created by the `kyma-gke-integration` job in a Google Cloud (GCP) project.

When the `kyma-gke-integration` job installs Kyma on the GKE cluster, GCP creates disk resources automatically.
Usually, the job that provisions the cluster cleans all such disks.
It can happen, however, that the job is terminated before its clean-up finishes.
This causes a resource leak that generates unwanted costs.
The garbage collector finds and removes such disks.

There are three conditions used to find disks for removal:
- The disk name pattern that is specific for the `kyma-gke-integration` job
- The disk users count where zero means the disk is unused
- The disk `creationTimestamp` value that is used to find disks existing at least for a preconfigured number of hours

Disks that meet all these conditions are subject to removal.

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
    --dryRun=false
```

### Flags

See the list of available flags:

| Name                      | Required | Description                                                                                          |
| :------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **--project**             |   Yes    | GCP project name
| **--dryRun**              |    No    | The boolean value that controls the dry-run mode. It defaults to `true`.
| **--ageInHours**          |    No    | The integer value for the number of hours. It only matches disks older than `now()-ageInHours`. It defaults to `2`.
| **--diskNameRegex**       |    No    | The string value with a valid Golang regexp. It is used to match disks by their name. It defaults to `^gke-gkeint.*[-]pvc[-]`.

### Environment Variables

See the list of available environment variables:

| Name                                  | Required | Description                                                                                          |
| :------------------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **GOOGLE_APPLICATION_CREDENTIALS**    |    Yes   | The path to the service account file. The service account requires at least `compute.zones.list`, `compute.disks.list`, and `compute.disks.delete` Google IAM permissions. |
