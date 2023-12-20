# Automated Approver

## Overview

With the Automated Approver tool, you can automatically approve pull requests based on the rules you define. The tool enables automation of the approval process for pull requests in repositories that need reviews before merge. The tool automates the PR review process without limiting user write permission on the repository. It can provide an automated review process for all pull request authors.

## How it works

![Automated Approver workflow](./assets/automated-approver.svg)

Automated approver is a Prow plugin written in Golang. GitHub events are dispatched by Prow to the Automated approver plugin. Automated approver runs in a Prow Kubernetes cluster along with Prow components.

Automated Approver reacts to the following events:
 - pull request review requested
 - pull request synchronized
 - review dismissed

To identify pull requests that must be approved by the tool, Automated Approver evaluates rules defined in a `rules.yaml` file. Rules are defined per organization, repository, or user entity. You can define the following conditions in the rules:
 - pull request required labels
 - pull request changed files

If a pull request meets the conditions, the tool checks if the pull request tests are finished. A `Tide` context is an exception: pending status for `Tide` is ignored. The tool uses a backoff algorithm for sleep duration between subsequent status checks. A `wait-for-statuses-timeout` flag defines a timeout period while waiting for statuses to finish and reports its state back to GitHub. Once the tests are finished, it checks whether they were successful. Currently, the tool doesn't support optional tests. When all checks and conditions are met, the tool approves the pull request.

Automated Approver uses the identity of a dedicated GitHub user to approve pull requests. Depending on repository configuration, the user must have write permission on the repository, must be added to repository collaborators, and as code owner in the `CODEOWNERS` file.

## How to use it

You configure Automated Approver with CLI flags. The flags are defined in the following files in our repository and their dependencies:
- Automated Approver [configuration flags](https://github.com/kyma-project/test-infra/blob/5242421660dab5979a763bcd596eba48bafe093d/cmd/external-plugins/automated-approver/main.go#L39) 
- External plugin [configuration flags](https://github.com/kyma-project/test-infra/blob/5242421660dab5979a763bcd596eba48bafe093d/pkg/prow/externalplugin/externalplugin.go#L68) 
Define the needed flags' values in the Pod specification and apply it to the Kubernetes cluster.

Additionally, Automated Approver uses rules to approve pull requests. You define the rules as a `yaml` file and apply them to the Kubernetes cluster as a config map. You must mount this config map to the Pod that runs Automated Approver.


## How to install it

Automated Approver runs in a Kubernetes cluster. A Pod and service specification is defined in the Kubernetes [deployment manifest file](../../../prow/cluster/components/automated-approver_external-plugin.yaml). A service is required for Prow to dispatch GitHub events to registered external plugins.

The rules against which Automated Approver validates pull requests are defined in a Kubernetes [config map manifest file](../../../configs/automated-approver-rules.yaml).

Automated Approver Kubernetes resources are managed by Terraform. Installation and updates are applied by running the `terraform apply` command automatically with our CI/CD system.
