[Rotate Gardener Service Account Secrets Using Cloud Run](/cmd/cloud-run/gardener-sa-rotate/README.md) - The Cloud Run application creates a new key for a GCP service account, updates the required secret data, and deletes old versions of a key. The function is triggered by a Pub/Sub message sent by a secret stored in Secret Manager.

[Rotate Service Account Secrets](/cmd/cloud-run/rotate-service-account/README.md) - RotateServiceAccount creates a new key for a Google Cloud service account and updates the required secret data. It's triggered by a  Pub/Sub message sent by a secret stored in Secret Manager. It runs as a cloud run container.

[Cleanup of Service Account Secrets](/cmd/cloud-run/service-account-keys-cleaner/README.md) - The Cloud Run service deletes old keys for a Google Cloud service account and updates the required secret data for all service account secrets stored in the Secret Manager. The service is triggered by a Cloud Scheduler job.

[Automated Approver](/cmd/external-plugins/automated-approver/README.md) - With the Automated Approver tool, you can automatically approve a pull request (PR) based on the rules you define. The tool enables automation of the approval process for PRs in repositories that need reviews before merge. The tool automates the PR review process without limiting user `write` permission on the repository. It can provide an automated review process for all PR authors.

[gardener-rotate](/cmd/gardener-rotate/README.md) - The gardener-rotate tool allows you to generate a new access token for the Gardener service accounts and update kubeconfig stored in Secret Manager.

[Image Builder](/cmd/image-builder/README.md) - Image Builder is a tool for building OCI-compliant images.

[Image Builder GitHub Workflow Integration](/cmd/image-builder/github-workflow-integration.md) - The Image Builder solution integrates with GitHub workflows and uses an Azure DevOps pipeline to run the process of building OCI

[Image Detector](/cmd/image-detector/README.md) - Image Detector is a tool for updating the security scanner config with the list of images in the Prow cluster. To achieve that, it receives paths to files used to deploy Prow or its components.

[image-syncer](/cmd/image-syncer/README.md) - image-syncer is used to copy container images from one registry to another.

[Image URL Helper](/cmd/image-url-helper/README.md) - Image URL Helper is a tool that provides the following subcommands:

[JobGuard ](/cmd/jobguard/README.md) - JobGuard is a simple tool that fetches all statuses for GitHub pull requests (PRs) and waits for some of them to finish.

[Clusters Garbage Collector](/cmd/tools/clusterscollector/README.md) - This command finds and removes orphaned clusters created by the `kyma-gke-integration` job in a Google Cloud project.

[Disks Garbage Collector](/cmd/tools/diskscollector/README.md) - This command finds and removes orphaned disks created by the `kyma-gke-integration` job in a Google Cloud project.

[IP Address and DNS Record Garbage Collector](/cmd/tools/dnscollector/README.md) - This command finds and removes orphaned IP Addresses and related DNS records created by GKE integration jobs in a Google Cloud project.

[External Secrets Checker](/cmd/tools/externalsecretschecker/README.md) - This command checks external Secrets synchronization status, and if every Secret has a corresponding external Secret.

[GCR Cleaner](/cmd/tools/gcrcleaner/README.md) - This command finds and removes old GCR images created by Jobs in the Google Cloud project.

[GitHub Release](/cmd/tools/githubrelease/README.md) - This command creates GitHub releases.

[IP cleaner](/cmd/tools/ipcleaner/README.md) - This command finds and removes orphaned IP addresses created by jobs in the Google Cloud project.

[Job Guard](/cmd/tools/jobguard/README.md) - Job Guard was moved to the [`cmd`](/cmd/jobguard) directory.

[Prow Job Tester](/cmd/tools/pjtester/README.md) - Prow Job tester is a tool for testing changes to the Prow Jobs' definitions and code running in Prow Jobs. It uses the production Prow instance to run chosen Prow Jobs with changes from pull requests (PRs) without going through multiple cycles of new PRs, reviews, and merges. The whole development can be done within one cycle.

[Render Templates](/cmd/tools/rendertemplates/README.md) - The Render Templates is a tool that reads the configuration from a [`config.yaml`](/templates/config.yaml) file and [`data`](/templates/data) files to generate output files, such as Prow component jobs. While the `config.yaml` file can hold configuration for an output file, you can place such data within the data files that hold configuration for related output files. Having separate files with grouped data is cleaner and easier to maintain than one huge config file.

[Virtual Machines Garbage Collector](/cmd/tools/vmscollector/README.md) - This command finds and removes orphaned virtual machines (VMs) created by Prow jobs in a Google Cloud project.

[Artifact Registry Creator Tool (GCP, Terraform)](/configs/terraform/modules/artifact-registry/README.md) - This is the GCP image registry creator tool. Use the registry to publish modules that should be accessible to internal SAP teams.

[Documentation Guidelines](/docs/documentation_guidelines.md) - Follow the rules listed in this document to provide high-quality documentation.

[Add a Custom Secret to Prow](/docs/how-to/how-to-add-custom-secret.md) - This tutorial shows how to add and use a custom secret in the Prow pipeline.

[Standard Terraform Configuration](/docs/how-to/how-to-create-standard-terraform-config.md) - This document describes the standard Terraform configuration that is used in the `test-infra` repository. 

[How to Name a Secret](/docs/how-to/how-to-name-secret.md) - This tutorial describes how to name a secret in Google Secret Manager.

[Docs](/docs/prow/README.md) - The folder contains documents that provide an insight into Prow configuration, development, and testing.

[Authorization](/docs/prow/authorization.md) - To deploy a Prow cluster, configure the following service accounts in the Google Cloud project you own.

[Crier](/docs/prow/crier.md) - Crier reports the Prow Job status changes. For now, it is responsible for Slack notifications as Plank is still reporting the Prow Job statuses to GitHub.

[Run K3d Cluster Inside ProwJobs](/docs/prow/k3d-no-ssh.md) - This document provides simple instructions, with examples, on how to prepare a ProwJob to use a K3d cluster and Docker.

[Manage Component Jobs with Templates](/docs/prow/manage-component-jobs-with-templates.md) - This document describes how to define, modify, and remove Prow jobs for Kyma components using predefined templates that create both presubmit and postsubmit jobs for your component. Also, this document gives you the steps required to prepare your component for the Prow CI pipeline.

[Obligatory Security Measures](/docs/prow/obligatory-security-measures.md) - Read about the obligatory security measures to take on a regular basis and when a Kyma organization member leaves the project.

[Run ProwJobs in KinD or k3d](/docs/prow/pj-in-kind.md) - This document provides brief instructions on how to run ProwJobs in local kind (Kubernetes-in-Docker) or k3d locally.

[Presets](/docs/prow/presets.md) - This document contains the list of all Presets available in the [`config.yaml`](/prow/config.yaml) file. Use them to define Prow Jobs for your components.

[Prow Architecture](/docs/prow/prow-architecture.md) - The document outlines Prow architecture and interconnections between different systems and components that are involved in it.

[Prow Cluster Update](/docs/prow/prow-cluster-update.md) - Updating a Prow cluster requires an improved Prow version. The Kubernetes Prow instance gets updated via a shell script. The shell script offers only a short list of the last pushed container tags and as a result, limits the versions to choose from. To cherry-pick updates, monitor [Prow announcements](https://docs.prow.k8s.io/docs/announcements/) to see when fixes or important changes are merged into the Kubernetes repository. This document describes how to update a Prow cluster using a cherry-picked Prow version.

[HTML Lens](/docs/prow/prow-html-lens.md) - Spyglass HTML lens allows to render HTML files in the job results.

[Image Autobump ](/docs/prow/prow-jobs-autobump.md) - This document provides an overview of autobump Prow Jobs. 

[Prow Jobs QuickStart](/docs/prow/prow-jobs-quick-start.md) - This document provides an overview of how to quickly start working with Prow jobs.

[Prow Cluster Monitoring Setup](/docs/prow/prow-monitoring.md) - This document describes how to install and manage Prow cluster monitoring. 

[Security Leaks Scanner](/docs/prow/security_commit_scanner.md) - Security Leaks Scanner is a tool that scans a repository for potential security leaks, thus providing protection against any potential security threats and vulnerabilities. It operates using [Gitleaks](https://github.com/zricethezav/gitleaks), which ensures a thorough and efficient examination of your repository. 

[Prow Test Clusters](/docs/prow/test-clusters.md) - This document gathers information about test clusters that Prow jobs build. All test clusters are built in the `sap-kyma-prow-workloads` project.

[Tide Introduction](/docs/prow/tide-introduction-notes.md) - Along with the Prow upgrade, we want to introduce Tide for merging the PRs automatically.

[Prow Workload Clusters](/docs/prow/workload-clusters.md) - This document describes workload clusters on which Prow schedules Pods to execute the logic of a given Prow job. All workload clusters are aggregated under the `kyma-prow` Google Cloud project. We use two workload clusters for trusted and untrusted Prow jobs.

[Changelog Generator](/experimental/changelog-generator/README.md) - This project is a Docker image that is used to generate a changelog in the `kyma` repository. It uses GitHub API to get pull requests (PRs) with specified labels.

[Prow Runtime Images](/images/README.md) - This directory contains images that can be used as runtime images for all ProwJobs in Kyma's Prow Instance.

[E2E DinD K3d Image](/images/e2e-dind-k3d/README.md) - This image contains common tools for all jobs/tasks that test Kyma modules in K3d.

[PR Tag Builder](/pkg/tools/prtagbuilder/README.md) - PR Tag Builder is a tool that finds a pull request (PR) number for a commit.

[Cluster](/prow/cluster/README.md) - This folder contains files related to the configuration of the Prow production cluster that are used during the cluster provisioning.

[Resources](/prow/cluster/resources/README.md) - This directory contains Helm charts used by a Prow cluster.

[External Secrets](/prow/cluster/resources/external-secrets/README.md) - Kubernetes Secrets are synchronized with Google Cloud Secret Manager using [External Secrets Operator](https://github.com/external-secrets/external-secrets).

[Images](/prow/images/README.md) - > DEPRECATED: Use the [`images`](/images) directory instead.

[Golangci-Lint Image](/prow/images/golangci-lint/README.md) - This folder contains the Golangci-lint image that is based on the upstream Golangci-lint image. Use it to lint Go source files.

[Vulnerability Scanner](/prow/images/whitesource-scanner/README.md) - This folder contains the WhiteSource Unified Agent image that is based on the Java Buildpack image. Use it to perform WhiteSource vulnerability scans.

[Templates](/templates/README.md) - Jobs and Prow configuration are generated from templates by the Render Templates tool. Check

