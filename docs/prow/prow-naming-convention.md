# Naming Conventions

The document focuses mainly on test Prow instance and its resources hosted in Google Cloud Platform.

This page describes naming convention standards for Prow resources as well as ProwJob resources like:
- Service accounts
- Cryptographic keys
- Storage buckets


## Service Accounts (SAs)

Each Service Account (SA) must follow standard naming convention that looks, as follows:
- `sa` - prefix for not private service accounts.
- `first_letter_of_a_name+last_name` - prefix for private service accounts i.e. `hpotter`.
Personal SAs are limited to the project scope and their names consist of two parts only: `first_letter_of_a_name+last_name` and `gcp-project`.
Personal SA examples: `fflinstone-prow`, `hpotter-workloads`

- `gcp-project`- one of GCP project shortcuts: `workloads`, `prow` or `dev`.

| Project name   | Shortcut |
| :-----------------| :---------------- | 
| sap-kyma-prow-workloads | workloads|
| sap-kyma-prow | prow|
| sap-kyma-prow-neighbors-dev | dev|


- `application-name` - application name, purpose or tool where SA is used (ie. SAs used by prow itself: `prow-plank`, `prow-kyma-artifacts` and one SA used by job `job-kyma-integration`)

Service account name examples:
`sa-prow-gcs-plank`
`sa-workloads-kyma-backup-restore`

For short-living and test resources created in `sap-kyma-prow-neighbors-dev` and `sap-kyma-prow-neighbors-workloads` project it is necessary to easily identify group of resources by adding `commit-sha` prefix.
`commit-sha` represents commit ID in the `/test-infra` repository where test pipeline was triggered. Commit ID is rendered automatically while creating or merging pull request.
Example SA name: `c177396-sa-prow-gcs-plank`.

### Additional remark

```
gcloud iam service-accounts create [SA-NAME]
--description "[SA-DESCRIPTION]"
--display-name "[SA-DISPLAY-NAME]"
```
While creating SA please note that `[SA-DISPLAY-NAME]` should be the same as `[SA-NAME]` parameter and`[SA-DESCRIPTION]` is meaningful.


## Key Management Service
In order to limit the scope of data accessible with any single key version each project should have one key ring per project.

| KEY RING         | KEY | PROJECT NAME           |
| ------------- |:-------------:|:-------------:|
| kyma-prow |  kyma-prow-encryption |sap-kyma-prow |
| prow-workloads | prow-workloads-encryption |sap-kyma-prow-workloads |
| neighbors-dev | neighbors-de-encryption |sap-kyma-neighbors-dev | 


## Storage buckets

This part of documentation concerns buckets created on `dev` or `workloads` project only. Guidelines related to the production Prow instance are [here](.production-cluster-configuration.md).

Short-living and test buckets, created in `sap-kyma-prow-neighbors-dev` and `sap-kyma-prow-neighbors-workloads` are prefixed with `commit-sha`, just like SAs.
Examples: `c177396-kyma-dev-logs`, `c177396-kyma-dev-secrets`