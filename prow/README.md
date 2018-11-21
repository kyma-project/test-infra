# Prow

## Overview

Prow is a Kubernetes-developed system that you can use as a Continuous Integration (CI) tool for validating your GitHub repositories and components, managing automatic validation of pull requests, applying and removing labels, or opening and closing issues.

You interact with Prow using slash (/) commands, such as `/test all`. You add them on pull requests or issues to trigger the predefined automation [plugins](https://status.build.kyma-project.io/plugins) that perform certain actions in respond to GitHub events. Upon proper configuration, GitHub events trigger jobs that are single-container Pods, created in dedicated builds and Kubernetes clusters by a microservice called Plank that is running in Google Cloud Platform (GCP). Each Prow component is a small Go service that has its own function in the management of these one-off single-pod Prow jobs.

In the context of the `kyma-project` organization, the main purpose of Prow is to serve as an external CI test tool that replaces the internal CI system.

### Basic configuration

Prow replies on this basic set of configurations:

- Kubernetes cluster deployed in Google Kubernetes Engine (GKE)
- GitHub bot account
- GitHub tokens:
  - `hmac-token` which is a Prow HMAC token used to validate GitHub webhooks
  - `oauth-token` which is a GitHub token with read and write access to the bot account
- Service accounts and their Secret files for sensitive jobs that are encrypted using Key Management Service (KMS) and stored in Google Cloud Storage (GCS)
- The `starter.yaml` file with a basic configuration of Prow components
- Webhooks configured for the GitHub repository to enable sending Events from a GitHub repository to Prow.
- Plugins enabled by creating and modifying the `plugins.yaml` file
- Jobs enabled by creating and configuring the basic `config.yaml` file, and specifying job definitions in the `jobs` subfolder

### Basic rules

Follow these basic rules when working with Prow in the `kyma-project` organization:

- You cannot test Prow configuration locally on Minikube. Perform all tests on the cluster.
- Avoid provisioning long-running clusters.
- Test Prow configuration against your `kyma` fork repository.
- Disable builds on the internal CI only after all CI functionalities are provided by Prow. This applies not only to the `master` branch but also to release branches.

### Project structure

The `prow` folder contains a set of configuration files for the Prow production cluster.

<!-- Update the folder structure each time you modify it. -->

Its structure looks as follows:

```

  ├── cluster               # Files for Prow cluster provisioning
  ├── images                # Images for Prow jobs
  ├── jobs                  # Files with job definitions
  ├── scripts               # Scripts used by the test jobs
  ├── config.yaml           # The main Prow configuration, without job definitions. For example, it contains Plank configuration and Preset definitions.
  ├── install-prow.sh       # The script for the production cluster installation
  └── plugins.yaml          # The file with Prow plugins configuration
```

## Installation

Read the [`docs`](../docs/prow/README.md) to lean how to configure the production Prow or install it on your forked repository for development and testing.

## Development

Read about the conventions for organizing and naming jobs in the `prow` subdirectories.

### Strategy for organizing jobs

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
| | | | |- environments.yaml
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

### Upload configuration to the production Prow cluster

Prow configuration is automatically uploaded to the production cluster from the `master` branch by the **Config Updater** plugin.

### Configure branch protection

Prow is responsible for setting branch protection on repositories. The configuration of branch protection is defined in `config.yaml`.

After you create a new job, define it as required for pull requests. Add the job context to the `required_status_checks.contexts` list in the proper repository.

See the sample configuration for the `test-infra` repository:

```yaml
branch-protection:
  orgs:
    kyma-project:
      repos:
        test-infra:
          enforce_admins: false
          required_pull_request_reviews:
            dismiss_stale_reviews: false
            require_code_owner_reviews: true
            required_approving_review_count: 1
          protect: true
          required_status_checks:
            contexts:
              - license/cla
              - prow/test-infra/validate-scripts
              - prow/test-infra/validate-configs
              - prow/test-infra/bootstrap
              - prow/test-infra/buildpack-golang
              - prow/test-infra/buildpack-node
              - prow/test-infra/cleaner
              - prow/test-infra/development/tools
              - prow/test-infra/test-jobs-yaml-definitions
```

The Branch Protector component updates the configuration every 30 minutes.
