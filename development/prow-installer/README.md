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

## Config

For an example of a configuration please see [prow-installer-config.yaml](https://github.com/kyma-project/test-infra/blob/master/development/prow-installer/config/prow-installer-config.yaml)

## About

In order to achieve this, the installer is broken up into several packages that each do their own thing.
- [cluster](https://github.com/kyma-project/test-infra/tree/master/development/prow-installer/pkg/cluster) is handling cluster creation/deletion
- [k8s](https://github.com/kyma-project/test-infra/tree/master/development/prow-installer/pkg/k8s) is providing a k8s client to a given cluster
- [secrets](https://github.com/kyma-project/test-infra/tree/master/development/prow-installer/pkg/secrets) is an interface to the GCP secrets api for managing encrypted keys & keyrings
- [roles](https://github.com/kyma-project/test-infra/tree/master/development/prow-installer/pkg/roles) is an interface for IAM role management
- [serviceaccount](https://github.com/kyma-project/test-infra/tree/master/development/prow-installer/pkg/serviceaccount) is an interface for IAM service account management
- [storage](https://github.com/kyma-project/test-infra/tree/master/development/prow-installer/pkg/storage) is an interface for bucket management
- [installer](https://github.com/kyma-project/test-infra/tree/master/development/prow-installer/pkg/installer) functionality for uninstall