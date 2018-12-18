# Clusters Garbage Collector

## Overview

This command finds and removes orphaned clusters created by the `kyma-gke-integration` job in a Google Cloud Platform (GCP) project.

The `kyma-gke-integration` job creates a GKE cluster to install and test Kyma.
Usually, the job also cleans up the cluster.
It can happen, however, that the job is terminated before its clean-up finishes.
This causes a resource leak that generates unwanted costs.
The garbage collector finds and removes such clusters.

There are three conditions used to find clusters for removal:
- The cluster name pattern that is specific for the `kyma-gke-integration` job
- The value of a `job` label the cluster is annotated with
- The cluster `createTime` value that is used to find clusters existing at least for a preconfigured number of hours

Clusters that meet these conditions are subject to removal.

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
| **--ageInHours**          |    No    | The integer value for the number of hours. It only matches clusters older than `now()-ageInHours`. It defaults to `3`.
| **--clusterNameRegexp**   |    No    | The string value with a valid Golang regexp. It is used to match clusters by their name. It defaults to `^gkeint[-](pr|commit)[-].*`.
| **--jobLabelRegexp**      |    No    | The string value with a valid Golang regexp. It is used to match clusters by the `job` label value. It defaults to `^kyma-gke-integration$`.

### Environment variables

See the list of available environment variables:

| Name                                  | Required | Description                                                                                          |
| :------------------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **GOOGLE_APPLICATION_CREDENTIALS**    |    Yes   | The path to the service account file. The service account requires at least `container.clusters.list` and `container.clusters.delete` Google IAM permissions. |

