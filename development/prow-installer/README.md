# prow-installer

## Usage

Run the following command to get an automatic installation of prow with the configuration from your config file:

```sh
go run cmd/installer/main.go --config <path to config> --credentials-file <path to GCP service account json>
```

> :exclamation: The tool will override `GOOGLE_APPLICATION_CREDENTIAL` environment variable to function :exclamation:

This will trigger in the following order (on GCP) creating from config:
- GKE Cluster(s)
- Storage Bucket(s)
- Service Account(s)

And then populate secrets into the created GKE Cluster(s)

Clusters will be labeled with a "created-at" timestamp label.

Additionally, the `--remove true` parameter will delete all resources created by a given config.