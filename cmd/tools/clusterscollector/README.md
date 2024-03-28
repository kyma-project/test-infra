# Clusters Garbage Collector

## Overview

This command finds and removes orphaned clusters created by the `kyma-gke-integration` job in a Google Cloud project.

The `kyma-gke-integration` job creates a GKE cluster to install and test Kyma.
Usually, the job also cleans up the cluster.
It can happen, however, that the job is terminated before its clean-up finishes.
This causes a resource leak that generates unwanted costs.
The garbage collector finds and removes such clusters based on two different strategies, specified by the caller.

In the `default` filter strategy, there are three conditions used to find clusters for removal:
- The cluster name pattern that is specific for the `kyma-gke-integration` job
- The value of a `volatile` label the cluster is annotated with
- The cluster `createTime` value that is used to find clusters existing at least for a preconfigured number of hours

In the `time` filter strategy, there are three conditions used to find clusters for removal:
- The label value of `volatile` that the cluster is annotated with
- The label value of `created-at`, which holds the unix timestamp of when the cluster was created
- The label value of `ttl`, which specifies the clusters maximum intended runtime in hours

Clusters that meet all these conditions in the specified strategy are subject to removal.

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
| **--project**             |   Yes    | Google Cloud project name
| **--dryRun**              |    No    | The boolean value that controls the dry-run mode. It defaults to `true`.
| **--strategy**            |    No    | The cluster filter strategy. Defaults to `default`, can be switched to `time`.
| **--ageInHours**          |    No    | The integer value for the number of hours. It only matches clusters older than `now()-ageInHours`. It defaults to `3`. [Only honored in `default` strategy]
| **--clusterNameRegexp**   |    No    | The string value with a valid Golang regexp. It is used to match clusters by their name. It defaults to `^gkeint[-](pr|commit)[-].*`. [Only honored in `default` strategy]
| **--excluded-clusters**   |    No    | The list of clusters that cannot be removed by the cluster collector.

### Environment Variables

See the list of available environment variables:

| Name                                  | Required | Description                                                                                          |
| :------------------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **GOOGLE_APPLICATION_CREDENTIALS**    |    Yes   | The path to the service account file. The service account requires at least `container.clusters.list` and `container.clusters.delete` Google IAM permissions. |

