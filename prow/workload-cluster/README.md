# Cluster

## Overview

This folder contains configuration files for the Prow workload. This configuration is used during cluster provisioning.

## Project structure

<!-- Update the folder structure each time you modify it. -->

The folder structure looks as follows:

```
  ├── 00-clusterrolebinding.yaml                # The enabled Prow cluster to access workload cluster and run all jobs.
  ├── 02-kube-system_poddisruptionbudgets.yaml  # The definition of Pod Disruption Budgets for Pods in the  `kube-system` Namespace, used to unblock the node autoscaler.
  └── required-secrets.yaml             # A default list of required Secrets that must be stored in a storage bucket.
```

## Required Secrets
The `secretspopulator` function reads the `required-secrets.yaml` file which includes required Secrets stored in a Google Cloud Storage (GCS) bucket.
You can define two kinds of Secrets:
- Service accounts
- Generic Secrets

The `Secretspopulator` looks for a `{prefix}.encrypted` object in a bucket and creates a Kubernetes Secret with a `{prefix}` name.
For service accounts, the Secret key is `service-account.json`. For generic Secrets, you must provide a key.
For more details about the syntax of this file, see the `RequiredSecretsData` type in `development/tools/cmd/secretspopulator/secretspopulator`.
