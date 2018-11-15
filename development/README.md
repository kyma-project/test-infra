# Development

## Overview

The purpose of the folder is to store scripts used for the development of the `test-infra` repository.

### Project structure

The `development` folder has the following structure:
<!-- Update the project structure each time you modify it. -->

```
  ├── checker                  # Code sources of a simple Go application that verifies the configuration of the "config.yaml" and "plugins.yaml" files.          
  ├── check.sh                 # The script that runs the "Checker" application.                                              
  ├── create-gcp-secrets.sh    # The script that downloads Secrets from the GCP storage bucket to your Prow installation.
  ├── install-prow.sh          # The script that installs Prow on your cluster.
  ├── provision-cluster.sh     # The script that creates a Kubernetes cluster.
  ├── remove-prow.sh           # The script that removes Prow from your Kubernetes cluster.
  ├── update-config.sh         # The script that updates the new configuration of the "config.yaml" file on a cluster.
  ├── update-plugins.sh        # The script that updates the new configuration of the "plugins.yaml" file on a cluster.
  └── validate-scripts.sh      # The script that performs a static analysis of bash scripts in the "test-infra" repository.

```
