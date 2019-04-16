# IP release

## Overview

This command finds and removes ips created by the `kyma-gke-long-lasting` job in a Google Cloud Platform (GCP) project.

The `kyma-gke-long-lasting` job creates a GKE cluster to install and test Kyma.

## Usage

For safety reasons, the dry-run mode is the default one.
To run it, use:
```bash
env GOOGLE_APPLICATION_CREDENTIALS={path to service account file} go run main.go \
    --project={gcloud project name} \
    --zone={gcloud zone} \
    --name={resource name} \
    --address={address of the resource}
```

To turn the dry-run mode off, use:
```bash
env GOOGLE_APPLICATION_CREDENTIALS={path to service account file} go run main.go \
    --project={gcloud project name} \
    --zone={gcloud zone} \
    --name={resource name} \
    --address={address of the resource} \
    --dryRun=false
```

### Flags

See the list of available flags:

| Name                      | Required | Description                                                                                          |
| :------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **--project**             |   Yes    | GCP project name
| **--zone**                |   Yes    | GCP zone name
| **--name**                |   Yes    | GCP DNS resource name
| **--address**             |   Yes    | GCP resource's attached IP
| **--rtype**               |    No    | DNS Record Type to search for, default: A
| **--ttl**                 |    No    | TTL of the resource, default: 300
| **--maxAttempts**         |    No    | Maximum number of retries in the backoff, default: 3
| **--backoff**             |    No    | Initial backoff in seconds for the first retry, will increase after this, default: 5
| **--dryRun**              |    No    | The boolean value that controls the dry-run mode. It defaults to `true`.

### Environment variables

See the list of available environment variables:

| Name                                  | Required | Description                                                                                          |
| :------------------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **GOOGLE_APPLICATION_CREDENTIALS**    |    Yes   | The path to the service account file. The service account requires at least `container.clusters.list` and `container.clusters.delete` Google IAM permissions. |

