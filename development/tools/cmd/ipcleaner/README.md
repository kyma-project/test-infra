# IP cleaner

## Overview

This command finds and removes IPs created by the `kyma-gke-long-lasting` job in a Google Cloud Platform (GCP) project.

The `kyma-gke-long-lasting` job creates a GKE cluster to install and test Kyma.

## Usage

For safety reasons, the dry-run mode is the default one.
To run it, use:
```bash
env GOOGLE_APPLICATION_CREDENTIALS={path to service account file} go run main.go \
    --project={gcloud project name} \
    --ipname={gcloud resource name} \
    --region={gcloud region}
```

To turn the dry-run mode off, use:
```bash
env GOOGLE_APPLICATION_CREDENTIALS={path to service account file} go run main.go \
    --project={gcloud project name} \
    --ipname={gcloud resource name} \
    --region={gcloud region} \
    --dryRun=false
```

### Flags

See the list of available flags:

| Name                      | Required | Description                                                                                          |
| :------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **--project**             |   YES    | GCP project name.
| **--ipname**              |   YES    | GCP IP resource name.
| **--region**              |   YES    | GCP region name.
| **--maxAttempts**         |    NO    | Maximum number of retries in the backoff. The default value is `3`.
| **--backoff**             |    NO    | Initial backoff in seconds for the first retry. The backoff will increase after this time. The default value is `5`.
| **--dryRun**              |    NO    | The boolean value that controls the dry-run mode. The default value is `true`.

### Environment variables

See the list of available environment variables:

| Name                                  | Required | Description                                                                                          |
| :------------------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **GOOGLE_APPLICATION_CREDENTIALS**    |    YES   | The path to the service account file. The service account requires at least `container.clusters.list` and `container.clusters.delete` Google IAM permissions. |

