# Cluster

## Overview

This folder contains files related to the configuration of the Prow production cluster that are used during the cluster provisioning.

### Project structure

<!-- Update the folder structure each time you modify it. -->

The structure of the folder looks as follows:

```
  ├── 00-deck-patch.yaml                # The patch file for the Deck Deployment used to enable HTTP to HTTPS redirection.
  ├── 01-cert-manager.yaml              # The definition of the Cert Manager and related resources, required to manage the SSL certificates that ensure the trusted website connection
  ├── 02-cluster-issuer.yaml            # The definition of the resource which creates new certificates
  ├── 03-tls-ing_ingress.yaml           # The definition of the encrypted Ingress that accesses the Prow cluster
  ├── 04-branchprotector_cronjob.yaml   # The definition of the Branch Protector CronJob that configures protection on branches
  ├── 05-tiller.yaml                    # The definition of the Tiller that is used for Helm integration
  ├── 06-pushgateway_deployment.yaml    # The definition of the Pushgateway that is used for monitoring
  ├── starter.yaml                      # The basic definition of Prow, including ConfigMaps, Deployments, and CustomResourceDefinitions
  └── required-secrets.yaml             # A default list of required Secrets that must be stored in a storage bucket
```

### Required secrets structure
The `required-secrets.yaml` file is read by `secretspopulator` and consists of required Secrets stored in a Google Cloud Storage (GCS) bucket.
You can define two kinds of Secrets:
- Service accounts
- Generic Secrets

`Secretspopulator` looks for a `{prefix}.encrypted` object in a bucket and creates a Kubernetes Secret with a `{prefix}` name.
For service accounts, the Secret key is `service-account.json`. For generic Secrets, you must provide a key.
For more details about the syntax of this file, see the `RequiredSecretsData` type in `development/tools/cmd/secretspopulator/secretspopulator`.
