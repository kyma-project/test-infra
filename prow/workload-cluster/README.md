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

Creating a cluster with `--security-group="gke-security-groups@sap.com` parameter allows you to apply set of custom privileges to the specific group of people.
You can ask Neighbors team to create custom Google Group containing certain members, additionally, the group itself has to be a member of the already created Google Group gke-security-groups@sap.com.
When it is done you can create Roles, ClusterRoles, RoleBindings, and ClusterRoleBindings that reference your G Suite Google Groups.

Kyma release cluster is an example where such configuration is used.