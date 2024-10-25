[Rotate Service Account Secrets](/cmd/cloud-run/rotate-service-account/README.md) - RotateServiceAccount creates a new key for a Google Cloud service account and updates the required secret data. It's triggered by a  Pub/Sub message sent by a secret stored in Secret Manager. It runs as a cloud run container.

[Cleanup of Service Account Secrets](/cmd/cloud-run/service-account-keys-cleaner/README.md) - The Cloud Run service deletes old keys for a Google Cloud service account and updates the required secret data for all service account secrets stored in the Secret Manager. The service is triggered by a Cloud Scheduler job.

[Automated Approver](/cmd/external-plugins/automated-approver/README.md) - With the Automated Approver tool, you can automatically approve a pull request (PR) based on the rules you define. The tool enables automation of the approval process for PRs in repositories that need reviews before merge. The tool automates the PR review process without limiting user `write` permission on the repository. It can provide an automated review process for all PR authors.

[Image Builder](/cmd/image-builder/README.md) - Image Builder is a tool for building OCI-compliant images in an SLC-29-compliant system from a GitHub workflow.

[Image Builder GitHub Workflow Integration](/cmd/image-builder/github-workflow-integration.md) - The Image Builder solution integrates with GitHub workflows and uses an Azure DevOps pipeline to run the process of building OCI

[Image Builder](/cmd/image-builder/image-builder.md) - Image Builder is a tool for building OCI-compliant images.

[Image Detector](/cmd/image-detector/README.md) - Image Detector is a tool for updating the security scanner config with the list of images in the Prow cluster. To achieve that, it receives paths to files used to deploy Prow or its components.

[image-syncer](/cmd/image-syncer/README.md) - Contents:

[Image URL Helper](/cmd/image-url-helper/README.md) - Image URL Helper is a tool that provides the following subcommands:

[JobGuard ](/cmd/jobguard/README.md) - JobGuard is a simple tool that fetches all statuses for GitHub pull requests (PRs) and waits for some of them to finish.

[OIDC Token Verifier](/cmd/oidc-token-verifier/README.md) - The OIDC Token Verifier is a command-line tool designed to validate the OIDC token and its claim values. It is primarily used in the

[External Secrets Checker](/cmd/tools/externalsecretschecker/README.md) - This command checks external Secrets synchronization status, and if every Secret has a corresponding external Secret.

[Job Guard](/cmd/tools/jobguard/README.md) - Job Guard was moved to the [`cmd`](/cmd/jobguard) directory.

[Artifact Registry Creator Tool (GCP, Terraform)](/configs/terraform/modules/artifact-registry/README.md) - This is the GCP image registry creator tool. Use the registry to publish modules that should be accessible to internal SAP teams.

[Documentation Guidelines](/docs/documentation_guidelines.md) - Follow the rules listed in this document to provide high-quality documentation.

[Add a Custom Secret to Prow](/docs/how-to/how-to-add-custom-secret.md) - This tutorial shows how to add and use a custom secret in the Prow pipeline.

[Standard Terraform Configuration](/docs/how-to/how-to-create-standard-terraform-config.md) - This document describes the standard Terraform configuration that is used in the `test-infra` repository. 

[Name a Secret](/docs/how-to/how-to-name-secret.md) - This tutorial describes how to name a secret in Google Secret Manager.

[Docs](/docs/prow/README.md) - The folder contains documents that provide an insight into Prow configuration, development, and testing.

[Authorization](/docs/prow/authorization.md) - To deploy a Prow cluster, configure the following service accounts in the Google Cloud project you own.

[Crier](/docs/prow/crier.md) - Crier reports the Prow Job status changes. For now, it is responsible for Slack notifications as Plank is still reporting the Prow Job statuses to GitHub.

[Obligatory Security Measures](/docs/prow/obligatory-security-measures.md) - Read about the obligatory security measures to take on a regular basis and when a Kyma organization member leaves the project.

[Run ProwJobs in KinD or k3d](/docs/prow/pj-in-kind.md) - This document provides brief instructions on how to run ProwJobs in local kind (Kubernetes-in-Docker) or k3d locally.

[Presets](/docs/prow/presets.md) - This document contains the list of all Presets available in the [`config.yaml`](/prow/config.yaml) file. Use them to define Prow Jobs for your components.

[Prow Architecture](/docs/prow/prow-architecture.md) - The document outlines Prow architecture and interconnections between different systems and components that are involved in it.

[Prow Cluster Update](/docs/prow/prow-cluster-update.md) - Updating a Prow cluster requires an improved Prow version. The Kubernetes Prow instance gets updated via a shell script. The shell script offers only a short list of the last pushed container tags and as a result, limits the versions to choose from. To cherry-pick updates, monitor [Prow announcements](https://docs.prow.k8s.io/docs/announcements/) to see when fixes or important changes are merged into the Kubernetes repository. This document describes how to update a Prow cluster using a cherry-picked Prow version.

[HTML Lens](/docs/prow/prow-html-lens.md) - Spyglass HTML lens allows to render HTML files in the job results.

[Image Autobump ](/docs/prow/prow-jobs-autobump.md) - This document provides an overview of autobump Prow Jobs. 

[Prow Cluster Monitoring Setup](/docs/prow/prow-monitoring.md) - This document describes how to install and manage Prow cluster monitoring. 

[Prow Test Clusters](/docs/prow/test-clusters.md) - This document gathers information about test clusters that Prow jobs build. All test clusters are built in the `sap-kyma-prow-workloads` project.

[Tide Introduction](/docs/prow/tide-introduction-notes.md) - Along with the Prow upgrade, we want to introduce Tide for merging the PRs automatically.

[Prow Workload Clusters](/docs/prow/workload-clusters.md) - This document describes workload clusters on which Prow schedules Pods to execute the logic of a given Prow job. All workload clusters are aggregated under the `kyma-prow` Google Cloud project. We use two workload clusters for trusted and untrusted Prow jobs.

[Prow Runtime Images](/images/README.md) - This directory contains images that can be used as runtime images for all ProwJobs in Kyma's Prow Instance.

[PR Tag Builder](/pkg/tools/prtagbuilder/README.md) - PR Tag Builder is a tool that finds a pull request (PR) number for a commit.

[Cluster](/prow/cluster/README.md) - This folder contains files related to the configuration of the Prow production cluster that are used during the cluster provisioning.

[Resources](/prow/cluster/resources/README.md) - This directory contains Helm charts used by a Prow cluster.

[External Secrets](/prow/cluster/resources/external-secrets/README.md) - Kubernetes Secrets are synchronized with Google Cloud Secret Manager using [External Secrets Operator](https://github.com/external-secrets/external-secrets).

