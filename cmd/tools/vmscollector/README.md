# Virtual Machines Garbage Collector

## Overview

This command finds and removes orphaned virtual machines (VMs) created by Prow jobs in a Google Cloud (GCP) project.

Prow jobs create a VM instance to install and test Kyma.
Usually, the job also cleans up the instance.
It can happen, however, that the job is terminated before its clean-up finishes.
This causes a resource leak that generates unwanted costs.
The garbage collector finds and removes such VM instances.

There are three conditions used to find instances for removal:
- The instance name is not caught by the exclude names regex.
- The value of the `job-name` label the instance is annotated with is not caught by the exclude labels regex.
- The instance `creationTimestamp` value that is used to find instance existing at least for a preconfigured number of hours.

VM instances that meet all these conditions are subject to removal.

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
| **--ageInHours**          |    No    | The integer value for the number of hours. It only matches VM instances older than `now()-ageInHours`. It defaults to `3`.
| **--vmNameRegexp**        |    No    | The string value with a valid Golang regex. It is used to exclude VM instances by their name. It defaults to `^gke-nightly-.*\|gke-weekly.*\|shoot--kyma-prow.*`.
| **--jobLabelRegexp**      |    No    | The string value with a valid Golang regex. It is used to exclude VM instances by the `job-name` label value. It defaults to `^kyma-gke-nightly\|kyma-gke-nightly-.*\|kyma-gke-weekly\|kyma-gke-weekly-.*$`.

### Environment Variables

See the list of available environment variables:

| Name                                  | Required | Description                                                                                          |
| :------------------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **GOOGLE_APPLICATION_CREDENTIALS**    |    Yes   | The path to the service account file. The service account requires at least `compute.instances.list` and `compute.instances.delete` Google IAM permissions. |
