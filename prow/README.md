# Prow

## Overview

Prow is a Kubernetes-developed system that you can use as a Continuous Integration (CI) tool for validating your GitHub repositories and components, managing automatic validation of PRs, applying and removing labels, or opening and closing issues.

You interact with Prow using slash (/) commands, such as `/test all`. You add them on pull requests or issues to trigger the predefined automation [plugins](https://prow.k8s.io/plugins) that perform certain actions in respond to GitHub events. Upon proper configuration, GitHub events trigger jobs that are single-container Pods, created in dedicated builds and Kubernetes clusters by a microservice called Plank that is running in GCP. Each Prow component is a small Go service that has its own function in the management of these one-off single-pod ProwJobs.

In the content of the `kyma-project` organization, the main purpose of Prow is to serve as an external CI test tool that replaces the internal CI system.

Prow configuration replies on this basic set of configurations:
- Kubernetes cluster deployed in Google Container Engine (GKE).
- GitHub bot account.
- GitHub tokens:
    - `hmac-token` which is a Prow HMAC token used for validating GitHub webhooks.
    - `oauth-token` which is a GitHub token with read and write access to the bot account.
- Service accounts and their Secret files for sensitive jobs that are encrypted using Key Management Service (KMS) and stored in Google Cloud Storage (GCS).
- The `starter.yaml` file with a basic configuration of Prow components.
- Webhooks configured for the GitHub repository to enable sending Events from a GitHub repository to Prow.
- Plugins enabled by creating and modifying the `plugins.yaml` file.
- Jobs enabled by creating and configuring the basic `config.yaml` file and additional `yaml` files for specific Kyma components.

### Basic rules

Follow these basic rules when working with Prow in the `kyma-project` organization:

- You cannot test Prow configuration locally on Minikube. Perform all the tests on the cluster.
- Avoid provisioning long-running clusters.
- Test Prow configuration against your `kyma` fork repository.
- Disable builds on the internal CI only after all CI functionalities are provided by Prow. This applies not only for the `master` branch but also for release branches.

### Project structure

The `prow` folder contains a set of configuration files for the Prow production cluster.

<!-- Update the folder structure each time you modify it. -->

Its structure looks as follows:

```

  ├── cluster               # All "yaml" files for Prow cluster provisioning.           
  ├── images                # Images for component jobs that you can also use for generic builds.                                             
  ├── jobs                  # All files with jobs definitions.
  ├── scripts               # Scripts used by the test jobs.
  ├── config.yaml           # The main Prow configuration, without job definitions. For example, it contains Plank configuration and Preset definitions.
  ├── install-prow.sh       # The script for the production cluster installation.
  └── plugins.yaml          # The file with Prow plugins configuration.
```

## Installation

Read the [`docs`](../docs/prow/README.md) to lean how to configure the production Prow or install it on your forked repository for development and testing.

## Development

Read about the conventions for organizing and naming jobs in the `prow` subdirectories.

### Strategy for organising jobs

The `jobs/{repository_name}` directories have subdirectories which represent each component and contain job definitions. Each file must have a unique name. Job definitions not connected to a particular component, like integration jobs, are defined directly under the `jobs/{repository_name}` directory.

For example:

   ```
   ...
   prow
   |- cluster
   | |- starter.yaml
   |- images
   |- jobs
   | |- kyma
   | | |- components
   | | | |- environments
   | | | | |- environments.jobs.yaml
   | | |- kyma.integration.yaml
   |- scripts
   |- config.yaml
   |- plugins.yaml
   ...
   ```

### Convention for naming jobs

When you define jobs for Prow, both **name** and **context** of the job must follow one of these patterns:

- `prow/{repository_name}/{component_name}/{job_name}` for components
- `prow/{repository_name}/{job_name}` for jobs not connected to a particular component

In both cases, `{job_name}` must reflect the job's responsibility.
