# Cluster

## Overview

This folder contains files related to the configuration of the Prow production cluster that are used during the cluster provisioning.

### Project structure

<!-- Update the folder structure each time you modify it. -->

The structure of the folder looks as follows:

```
  ├── 01-cert-manager.yaml        # Definition of the Cert Manager and related resources, required to manage the SSL certificates that ensure the trusted website connection.
  ├── 02-cluster-issuer.yaml      # Definition of the resource responsible for creating new certificates.
  ├── 03-tls-ing_ingress.yaml     # Definition of the encrypted Ingress that accesses the Prow cluster.
  └── starter.yaml                # Basic definition of Prow, including ConfigMaps, deployments, and Custom Resource definitions.
```
