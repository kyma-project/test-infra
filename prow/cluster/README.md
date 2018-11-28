# Cluster

## Overview

This folder contains files related to the configuration of the Prow production cluster that are used during the cluster provisioning.

### Project structure

<!-- Update the folder structure each time you modify it. -->

The structure of the folder looks as follows:

```
  ├── 01-cert-manager.yaml              # The definition of the Cert Manager and related resources, required to manage the SSL certificates that ensure the trusted website connection
  ├── 02-cluster-issuer.yaml            # The definition of the resource which creates new certificates
  ├── 03-tls-ing_ingress.yaml           # The definition of the encrypted Ingress that accesses the Prow cluster
  ├── 04-branchprotector_cronjob.yaml   # The definition of the Branch Protector CronJob that configures protection on branches
  ├── starter.yaml                      # The basic definition of Prow, including ConfigMaps, Deployments, and CustomResourceDefinitions
  └── required-secrets.yaml             # Define list of required secrets that needs to be stored in a Storage Bucket
```

#### Required secrets structure
The `required-secrets.yaml` file is read by `secretspopulator` and consists of required secrets stored in a Gcloud storage bucket.
You can define 2 kind of secrets:
- service accounts
- generic secrets

`Secretspopulator` looks for object `{prefix}.encrypted` in a bucket and creates Kubernetes secret with name `{prefix}`.
For service accounts, secret key is `service-account.json`, for generic secret you have to provide key.
For more details about structure, see type `RequiredSecretsData` in `development/tools/cmd/secretspopulator/secretspopulator`
