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

## Configuring cluster to use Google Groups

[Google Groups](https://groups.google.com/a/sap.com/forum/#!overview) give you possibility to gather Kyma Developers accounts and manage GCP Project permissions based on the group name. Additionally, you can grant Kubernetes Roles, ClusterRoles, RoleBindings, and ClusterRoleBindings to the specific Google Group on your cluster.

For example on Kyma release cluster `cluster-admin` ClusterRole is granted to all mambers of kyma_developers@sap.group group. 

- First, there was kyma_developers@sap.group created and added as a member of gke-security-groups@sap.com.

- In the next step release cluster was build with [--security-group="gke-security-groups@sap.com](https://github.com/kyma-project/test-infra/blob/7b84900e56679fccfbc9e6839a85ade1dabe72bd/prow/scripts/cluster-integration/helpers/provision-gke-cluster.sh#L60) parameter. 

- The last step was to create ClusterRoleBindings for the custom group, in this case it was kyma_developers@sap.com.

  ```
  kubectl create clusterrolebinding kyma-developers-group-binding --clusterrole="cluster-admin" --group="kyma_developers@sap.com"
  ```

You can ask Neighbors team to create new G Suite Google Group in sap.com  domain. In the next step it is necessary to add these groups to the membership of gke-security-groups@[yourdomain.com].