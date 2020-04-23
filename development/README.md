# Development

## Overview

The purpose of the folder is to store tools developed and used in the `test-infra` repository.

### Project structure

The `development` folder has the following structure:

<!-- Update the project structure each time you modify it. -->

```
  ├── checker                  # This folder contains code sources of a simple Go application that verifies the configuration of jobs, "config.yaml", and "plugins.yaml".
  ├── tools                    # This folder contains code sources of Go applications for the "test-infra" repository.
  ├── install-prow.sh          # This script installs Prow on your cluster.
  ├── provision-cluster.sh     # This script creates a Kubernetes cluster.
  ├── remove-prow.sh           # This script removes Prow from your Kubernetes cluster.
  ├── update-config.sh         # This script updates the new configuration of the "config.yaml" file on a cluster.
  ├── update-jobs.sh           # This script updates the new configuration of the jobs on a cluster.
  ├── update-plugins.sh        # This script updates the new configuration of the "plugins.yaml" file on a cluster.
  ├── validate-config.sh       # This script runs the "Checker" application and checks the uniqueness of jobs names.
  ├── validate-scripts.sh      # This script performs a static analysis of bash scripts in the "test-infra" repository.
  ├── clusters-cleanup.sh      # This script invokes the tool for cleaning orphaned clusters created by the "kyma-gke-integration" job.
  ├── vms-cleanup.sh           # This script invokes the tool for cleaning orphaned VM instances created by the "kyma-gke-integration" job.
  ├── disks-cleanup.sh         # This script invokes the tool for cleaning orphaned disks created by the "kyma-gke-integration" job.
  ├── loadbalancer-cleanup.sh  # This script invokes the tool for cleaning orphaned load balancers created by the "kyma-gke-integration" job.
  ├── firewall-cleanup.sh      # This script invokes the tool for cleaning orphaned firewall rules.
  ├── resources-cleanup.sh     # This script is a generic resource cleanup tool launcher.
  └── jobguard                 # This folder contains the jobguard tool source code. It's used to control dependencies between running jobs.

```

Every directory which holds a go module must follow the structure described in the [project layout](https://github.com/golang-standards/project-layout).
