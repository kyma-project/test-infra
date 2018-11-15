# Development

## Overview

The purpose of the folder is to store scripts used for the development of the `test-infra` repository.

### Project structure

The `development` folder has the following structure:
<!-- Update the project structure each time you modify it. -->

```
  ├── checker                  # This folder contains code sources of a simple Go application that verifies the configuration of the "config.yaml" and "plugins.yaml" files          
  ├── check.sh                 # This script runs the "Checker" application.  .                                            
  ├── create-gcp-secrets.sh    # This script downloads Secrets from the GCP storage bucket to your Prow installation.
  ├── install-prow.sh          # This script installs Prow on your cluster.
  ├── provision-cluster.sh     # This script creates a Kubernetes cluster.
  ├── remove-prow.sh           # This script removes Prow from your Kubernetes cluster.
  ├── update-config.sh         # This script updates the new configuration of the "config.yaml" file on a cluster.
  ├── update-plugins.sh        # This script updates the new configuration of the "plugins.yaml" file on a cluster.
  └── validate-scripts.sh      # This script performs a static analysis of bash scripts in the "test-infra" repository.

```
