# Virtual Machines Garbage Collector

## Overview

This command finds and removes orphaned virtual machines created by the `kyma-integration` job in a Google Cloud Platform (GCP) project.

The `kyma-integration` job creates a Virtual Machine (VM) instance to install and test Kyma.
Usually, the job also cleans up the instance.
It can happen, however, that the job is terminated before its clean-up finishes.
This causes a resource leak that generates unwanted costs.
The garbage collector finds and removes such VM instances.

There are three conditions used to find instances for removal:
- The instance name pattern that is specific for the `kyma-integration` job
- The value of a `job-name` label the instance is annotated with
- The instance `creationTimestamp` value that is used to find instance existing at least for a preconfigured number of hours

VM instances that meet these conditions are subject to removal.

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
| **--project**             |   Yes    | GCP project name.
| **--dryRun**              |    No    | The boolean value that controls the dry-run mode. It defaults to `true`.
| **--ageInHours**          |    No    | The integer value for the number of hours. It only matches VM instances older than `now()-ageInHours`. It defaults to `3`.
| **--vmNameRegexp**        |    No    | The string value with a valid Golang regexp. It is used to match VM instances by their name. It defaults to `^kyma-integration-test-.*`.
| **--jobLabelRegexp**      |    No    | The string value with a valid Golang regexp. It is used to match VM instances by the `job-name` label value. It defaults to `^kyma-integration$`.

### Environment variables

See the list of available environment variables:

| Name                                  | Required | Description                                                                                          |
| :------------------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **GOOGLE_APPLICATION_CREDENTIALS**    |    Yes   | The path to the service account file. The service account requires at least `compute.instances.list` and `compute.instances.delete` Google IAM permissions. |

