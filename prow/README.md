# Prow

## Overview

Prow is a Kubernetes-developed system that you can use as a Continuous Integration (CI) tool for validating your GitHub repositories and components, managing automatic validation of pull requests (PRs), applying and removing labels, or opening and closing issues.

You interact with Prow using slash (/) commands, such as `/test all`. You add them on PRs or issues to trigger the predefined automation [plugins](https://status.build.kyma-project.io/plugins) that perform certain actions in respond to GitHub events. Upon proper configuration, GitHub events trigger jobs that are single-container Pods, created in dedicated builds and Kubernetes clusters by a microservice called Plank that is running in Google Cloud Platform (GCP). Each Prow component is a small Go service that has its own function in the management of Prow jobs.

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
- Webhooks configured for the GitHub repository to enable sending events from a GitHub repository to Prow.
- Plugins enabled by creating and modifying the `plugins.yaml` file
- Jobs enabled by creating and configuring the basic `config.yaml` file, and specifying job definitions in the `jobs` subfolder

### Basic rules

Follow these basic rules when working with Prow in the `kyma-project` organization:

- You cannot test Prow configuration locally on Minikube. Perform all tests on the cluster.
- Avoid provisioning long-running clusters.
- Test Prow configuration against your forked `kyma` repository.
- Disable builds on the internal CI only after all CI functionalities are provided by Prow. This applies not only to the `master` branch but also to release branches.

### Project structure

The `prow` folder contains a set of configuration files for the Prow production cluster.

<!-- Update the folder structure each time you modify it. -->

Its structure looks as follows:

```

  ├── branding              # Files related to Kyma branding for a Prow cluster
  ├── cluster               # Files for Prow cluster provisioning
  ├── images                # Images for Prow jobs
  ├── jobs                  # Files with job definitions
  ├── scripts               # Scripts used by the test jobs
  ├── config.yaml           # The main Prow configuration without job definitions. For example, it contains Plank configuration and Preset definitions.
  ├── install-prow.sh       # The script which installs the production cluster
  └── plugins.yaml          # The file with Prow plugins configuration
  └── create-secrets-for-workload-cluster.sh  # The script which creates a Secret in the Prow cluster, used to access the workload cluster
  └── set-up-workload-cluster.sh  # The script which prepares the workload cluster to be used by the Prow cluster
```

## Installation

Read the [`docs`](../docs/prow/README.md) to learn how to configure the production Prow or install it on your forked repository for development and testing.

## Development

### Prow jobs

Read [this](../docs/prow/prow-jobs.md) document to learn more about Prow job definitions.

### Upload configuration to the production Prow cluster

Prow configuration is automatically uploaded to the production cluster from the `master` branch by the **Config Updater** plugin.

### Configure branch protection

Prow is responsible for setting branch protection on repositories. The configuration of branch protection is defined in `config.yaml`.

After you create a new job, define it as required for PRs. Add the job context to the `required_status_checks.contexts` list in the proper repository.

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
              - test-infra-validate-scripts
              - test-infra-validate-configs
              - test-infra-bootstrap
              - test-infra-buildpack-golang
              - test-infra-buildpack-node
              - test-infra-cleaner
              - test-infra-development/tools
              - test-infra-test-jobs-yaml-definitions
```

The Branch Protector component updates the configuration every 30 minutes.


### Test changes in scripts

If you modify scripts in the `test-infra` repository and you want to test the changes made, follow one of these scenarios:

- Create a PR with your changes and wait for the existing Prow jobs to verify your code.

> **NOTE**: This scenario works only if you modify the existing code, and requires a PR for every consecutive change.

- Add the **extra_refs** field to your Prow job and work directly on your branch. This pulls the repository and branch you chose into the job and executes the code from that location.

> **NOTE**: Remember to revert your changes after your merge the code.

```yaml
extra_refs:
  - org: {username}                 # Your GitHub username in the organisation
    repo: test-infra                # Your GitHub repository
    base_ref: dex-github              # Branch, tag, and release to use
    path_alias: github.com/kyma-project/test-infra  # Path to the location where you want to clone the code
```
