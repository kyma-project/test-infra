# Naming conventions

This document describes the naming conventions for the Prow test instance and its resources hosted in Google Cloud Platform.

More specifically, it provides the standards for naming resources used by Prow and ProwJob resources like:
- Service accounts
- Cryptographic keys
- Storage buckets


## Service accounts

Each service account (SA) must follow the standard naming convention that looks as follows:
- `sa` - the prefix for non-private SAs
- `{FIRST_LATTER_OF_YOUR_NAME_AND_FULL_LAST_NAME}` - the prefix for private SAs, such as `hpotter`.
Private SAs are limited to the project scope and their names only consist of two parts, the prefix and `gcp-project`.

- `gcp-project`- one of these GCP project's shortcuts: `workloads`, `prow`, or `dev`.

    | Project name   | Shortcut |
    | :-----------------| :---------------- | 
    | sap-kyma-prow-workloads | workloads|
    | sap-kyma-prow | prow|
    | sap-kyma-prow-neighbors-dev | dev|
The examples of such private SAs are `fflinstone-prow` or `hpotter-workloads`.
- `{APPLICATION_NAME}` - the name of the application or tool in which the SA is used, or the account purpose. For example, the names of SAs used by Prow are `prow-plank` or `prow-kyma-artifacts`, while the SA used by a job is named `job-kyma-integration`.

See these SA name examples:
- `sa-prow-gcs-plank`
- `sa-workloads-kyma-backup-restore`

For short-term and test resources created in the `sap-kyma-prow-neighbors-dev` and `sap-kyma-prow-neighbors-workloads` projects it is necessary to easily identify a group of resources by adding the `commit-sha` prefix.

> **NOTE:** In the future, `commit-sha` will represent a commit ID from the `/test-infra` repository in which the test pipeline is triggered. The commit ID is rendered automatically while creating or merging a pull request.

The example of such an SA name is `c177396-sa-prow-gcs-plank`.

> **CAUTION:** While creating an SA, note that `{SA_DISPLAY_NAME}` should be the same as the `{SA_NAME}` parameter, and `{SA_DESCRIPTION}` is meaningful.

```
gcloud iam service-accounts create {SA_NAME}
--description "{SA_DESCRIPTION}"
--display-name "{SA_DISPLAY_NAME}"
```

## Key Management Service

To limit the scope of data accessible with any key version, each project must have one key ring per project.

| KEY RING         | KEY | PROJECT NAME           |
| ------------- |:-------------:|:-------------:|
| kyma-prow |  kyma-prow-encryption |sap-kyma-prow |
| prow-workloads | prow-workloads-encryption |sap-kyma-prow-workloads |
| neighbors-dev | neighbors-de-encryption |sap-kyma-neighbors-dev | 


## Storage buckets

This section only refers to buckets created either on the `dev` or `workloads` project. You can find guidelines related to the production Prow instance [here](./production-cluster-configuration.md).

Short-term and test buckets created in the `sap-kyma-prow-neighbors-dev` and `sap-kyma-prow-neighbors-workloads` are prefixed with `commit-sha`, just like SAs.
The examples of such bucket names are `c177396-kyma-dev-logs` and `c177396-kyma-dev-secrets`.
