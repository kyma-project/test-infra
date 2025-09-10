[Rotate Service Account Secrets](/cmd/cloud-run/rotate-service-account/README.md) - RotateServiceAccount creates a new key for a Google Cloud service account and updates the required secret data. It's triggered by a  Pub/Sub message sent by a secret stored in Secret Manager. It runs as a cloud run container.

[Cleanup of Service Account Secrets](/cmd/cloud-run/service-account-keys-cleaner/README.md) - The Cloud Run service deletes old keys for a Google Cloud service account and updates the required secret data for all service account secrets stored in the Secret Manager. The service is triggered by a Cloud Scheduler job.

[Image Autobumper](/cmd/image-autobumper/README.md) - Image Autobumper is a tool for automatically updating the version of a Docker image in a GitHub repository.

[Image Builder](/cmd/image-builder/README.md) - Image Builder is a tool for building OCI-compliant images in an SLC-29-compliant system from a GitHub workflow.

[Image Builder GitHub Workflow Integration](/cmd/image-builder/github-workflow-integration.md) - The Image Builder solution integrates with GitHub workflows and uses an Azure DevOps pipeline to run the process of building OCI

[Image Builder: Maintenance Guide](/cmd/image-builder/image-builder.md) - Image Builder is a tool for building OCI-compliant images using the Azure DevOps (ADO) pipeline backend.

[Image Detector](/cmd/image-detector/README.md) - Image Detector is a tool for updating the security scanner config with the list of images in the Prow cluster. To achieve that, it receives paths to files used to deploy Prow or its components.

[image-syncer](/cmd/image-syncer/README.md) - Contents:

[Image URL Helper](/cmd/image-url-helper/README.md) - Image URL Helper is a tool that provides the following subcommands:

[OIDC Token Verifier](/cmd/oidc-token-verifier/README.md) - The OIDC Token Verifier is a command-line tool designed to validate the OIDC token and its claim values. It is primarily used in the

[Artifact Registry Module](/configs/terraform/modules/artifact-registry/README.md) - The Artifact Registry module for Google Cloud is designed to maintain a standardized and reusable way of creating Artifact Registry in Google Cloud.

[Documentation Guidelines](/docs/documentation_guidelines.md) - Follow the rules listed in this document to provide high-quality documentation.

[IaC Configuration Guideline](/docs/guidelines_iac.md) - This document outlines the standard Terraform configuration and provides guidelines.

[Guidelines for Managing Follow-Up Issues](/docs/how-to/how-to-manage-follow-up-issues.md) - This document provides clear guidelines on how to effectively handle follow-up issues identified during events such as Technical Sprint Reviews.

[Name a Secret](/docs/how-to/how-to-name-secret.md) - This tutorial describes how to name a secret in Google Secret Manager.

[Manage Workflow Controllers](/docs/how-to/how-to_manage_workflow_controller.md) - This guide explains how to manage Workflow Controllers. Workflow Controllers are responsible for orchestrating and triggering jobs in GitHub Actions workflows, especially for advanced scenarios like merge queues.

[Workflow Controller](/docs/what-is/what-is_workflow_controller.md) - Workflow Controller is a GitHub Actions workflow that orchestrates and triggers downstream workflows based on repository changes. It implements advanced CI/CD logic, such as merge queues and selective job execution, by applying path-based filters to changed files. It was created due to a lack of filtering capabilities in GitHub Actions workflows for merge queues.

