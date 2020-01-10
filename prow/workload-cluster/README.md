# Cluster

## Overview

This folder contains configuration files for the Prow workload. This configuration is used during cluster provisioning.

## Project structure

<!-- Update the folder structure each time you modify it. -->

The folder structure looks as follows:

```
  ├── 00-clusterrolebinding.yaml                # The enabled Prow cluster to access workload cluster and run all jobs
  ├── 02-kube-system_poddisruptionbudgets.yaml  # The definition of Pod Disruption Budgets for Pods in the `kube-system` Namespace used to unblock the Node autoscaler
  └── required-secrets.yaml                     # The default list of required Secrets that must be stored in a storage bucket
```

## Required Secrets
The `secretspopulator` function reads the `required-secrets.yaml` file which includes required Secrets stored in the Google Cloud Storage (GCS) bucket.
You can define two kinds of Secrets:
- Service accounts
- Generic Secrets

The `secretspopulator` function looks for the `{prefix}.encrypted` object in the bucket and creates a Kubernetes Secret with a `{prefix}`.
For service accounts, the Secret key is `service-account.json`. For generic Secrets, you must provide a key.
For details on the file syntax, see the `RequiredSecretsData` type in [`secretspopulator`](../../development/tools/cmd/secretspopulator/main.go).

## Configure cluster to use Google Groups

Previously, you could only grant roles to Google Cloud user accounts or Cloud IAM service accounts. Google Groups for GKE (Beta) now allows you to grant roles to the members of a G Suite Google Group. With this mechanism, the users and groups themselves are maintained by your G Suite administrators, completely outside of Kubernetes or Cloud Console.

Google Groups give you the possibility to gather Kyma Developers accounts and manage GCP Project permissions based on the group name. Additionally, you can grant Kubernetes Roles, ClusterRoles, RoleBindings, and ClusterRoleBindings to the specific Google Group on your cluster.

For example, all members of the `kyma_developers@sap.group` group receive the **cluster-admin** ClusterRole on the Kyma release cluster. The process used to look as follows:

1. System administrators created `kyma_developers@sap.group` and added it as a member of `gke-security-groups@sap.com`.
  ![dashboards](/docs/prow/assets/GGroups.png)

2. They built the release cluster with the [**--security-group="gke-security-groups@sap.com**](https://github.com/kyma-project/test-infra/blob/7b84900e56679fccfbc9e6839a85ade1dabe72bd/prow/scripts/cluster-integration/helpers/provision-gke-cluster.sh#L60) parameter. 

3. They created ClusterRoleBindings for the `kyma_developers@sap.com` custom group.

  ```
  kubectl create clusterrolebinding kyma-developers-group-binding --clusterrole="cluster-admin" --group="kyma_developers@sap.com"
  ```

If you want to leverage this solution, ask the [Neighbors](https://github.com/orgs/kyma-project/teams/prow/members?utf8=%E2%9C%93&query=role%3Amaintainer) team to create a new G Suite Google Group in the `sap.com` domain. The next step is to add this group as a member of `gke-security-groups@{yourdomain.com}`.
