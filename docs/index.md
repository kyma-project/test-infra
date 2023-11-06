[Rotate Gardener service account secrets using Cloud Run](/cmd/cloud-run/gardener-sa-rotate/README.md) - The Cloud Run application creates a new key for a GCP service account, updates the required secret data, and deletes old versions of a key. The function is triggered by a Pub/Sub message sent by a secret stored in Secret Manager.

[Rotate KMS secrets using Cloud Run](/cmd/cloud-run/kms-rotate/README.md) - The Cloud Run application decrypts and encrypts files in a bucket with the latest version of a KMS key, and deletes old versions of a key. The function is triggered by a HTTP POST request sent by a Cloud Scheduler.

[Rotate service account secrets](/cmd/cloud-run/rotate-service-account/README.md) - RotateServiceAccount creates a new key for a GCP service account and updates the required secret data. It's triggered by a  Pub/Sub message sent by a secret stored in Secret Manager. It runs as a cloud run container.

[Cleanup of service account secrets](/cmd/cloud-run/service-account-keys-cleaner/README.md) - The Cloud Run service deletes old keys for a GCP service account and updates the required secret data for all service account secrets stored in the Secret Manager. The service is triggered by a Cloud Scheduler job.

[Automated Approver](/cmd/external-plugins/automated-approver/README.md) - With the Automated Approver tool, you can automatically approve pull requests based on the rules you define. The tool enables automation of the approval process for pull requests in repositories that need reviews before merge. The tool automates the PR review process without limiting user write permission on the repository. It can provide an automated review process for all pull request authors.

[gardener-rotate](/cmd/gardener-rotate/README.md) - The gardener-rotate tool allows you to generate a new access token for the Gardener service accounts and update kubeconfig stored in the Secret Manager.

[image-builder](/cmd/image-builder/README.md) - This tool serves as an intelligent wrapper for `kaniko-project/executor`. It reduces the complexity of building Docker images and removes the need of using Docker in Docker when building images in K8s infrastructure.

[Image Detector](/cmd/image-detector/README.md) - Image Detector is a tool for updating the security scanner config with the list of images in the Prow cluster. To achieve that, it receives paths to files used to deploy Prow or its components.

[image-syncer](/cmd/image-syncer/README.md) - image-syncer is used to copy container images from one registry to another.

[Image URL Helper](/cmd/image-url-helper/README.md) - Image URL Helper is a tool that provides the following subcommands:

[JobGuard ](/cmd/jobguard/README.md) - JobGuard is a simple tool that fetches all statuses for GitHub pull requests and waits for some of them to finish.

[prowjobparser](/cmd/prowjobparser/README.md) - The prowjobparser is a helper tool which parses all Prow Jobs under the provided path, matches them against the provided label filters, and prints matching Prow Job names to the standard output.

[Clusters Garbage Collector](/cmd/tools/clusterscollector/README.md) - This command finds and removes orphaned clusters created by the `kyma-gke-integration` job in a Google Cloud Platform (GCP) project.

[Disks Garbage Collector](/cmd/tools/diskscollector/README.md) - This command finds and removes orphaned disks created by the `kyma-gke-integration` job in a Google Cloud Platform (GCP) project.

[IP Address and DNS Record Garbage Collector](/cmd/tools/dnscollector/README.md) - This command finds and removes orphaned IP Addresses and related DNS records created by GKE integration jobs in a Google Cloud Platform (GCP) project.

[External Secrets Checker](/cmd/tools/externalsecretschecker/README.md) - This command checks external Secrets synchronization status, and if every Secret has a corresponding external Secret.

[GCR cleaner](/cmd/tools/gcrcleaner/README.md) - This command finds and removes old GCR images created by Jobs in the Google Cloud Platform (GCP) project.

[Github issues](/cmd/tools/githubissues/README.md) - This command queries all open Github issues in an organization or repository, and loads that data to a BigQuery table.

[GitHub release](/cmd/tools/githubrelease/README.md) - This command creates GitHub releases.

[IP cleaner](/cmd/tools/ipcleaner/README.md) - This command finds and removes orphaned IP addresses created by jobs in the Google Cloud Platform (GCP) project.

[Job Guard](/cmd/tools/jobguard/README.md) - Job Guard was moved [here](/cmd/jobguard).

[Prow Job tester](/cmd/tools/pjtester/README.md) - Prow Job tester is a tool for testing changes to the Prow Jobs' definitions and code running in Prow Jobs. It uses the production Prow instance to run chosen Prow Jobs with changes from pull requests (PRs) without going through multiple cycles of new PRs, reviews, and merges. The whole development can be done within one cycle.

[Render Templates](/cmd/tools/rendertemplates/README.md) - The Render Templates is a tool that reads the configuration from a [`config.yaml`](/templates/config.yaml) file and [`data`](/templates/data) files to generate output files, such as Prow component jobs. While the `config.yaml` file can hold configuration for an output file, you can place such data within the data files that hold configuration for related output files. Having separate files with grouped data is cleaner and easier to maintain than one huge config file.

[Virtual Machines Garbage Collector](/cmd/tools/vmscollector/README.md) - This command finds and removes orphaned virtual machines (VMs) created by Prow jobs in a Google Cloud Platform (GCP) project.

[Artifact Registry creator tool (GCP, Terraform)](/configs/terraform/modules/artifact-registry/README.md) - This is the GCP image registry creator tool. Use the registry to publish modules that should be accessible to internal SAP teams.

[Documentation guidelines](/docs/documentation_guidelines.md) - Follow the rules listed in this document to provide high-quality documentation.

[Documentation](/docs/how-to/ado/README.md) - The folder contains documents that provide an insight into Azure DevOps (ADO) configuration, development, and testing.

[GitHub.com and Azure Pipeline (ADO) integration](/docs/how-to/ado/how-to_integrate-githubcom-azurepipeline.md) - We have to verify that the integration of an SAP Azure DevopsPipeline as a quality gate for github.com works.

[Add custom secret to Prow](/docs/how-to/how-to-add-custom-secret.md) - This tutorial shows how to add a custom secret and use it in the Prow pipeline.

[Standard Terraform configuration](/docs/how-to/how-to-create-standard-terraform-config.md) - This document describes the standard Terraform configuration that is used in `test-infra` repository. 

[Docs](/docs/prow/README.md) - The folder contains documents that provide an insight into Prow configuration, development, and testing.

[Authorization](/docs/prow/authorization.md) - To deploy a Prow cluster, configure the following service accounts in the GCP project you own.

[Crier](/docs/prow/crier.md) - Crier reports the Prow Job status changes. For now, it is responsible for Slack notifications as Plank is still reporting the Prow Job statuses to GitHub.

[Run K3d cluster inside ProwJobs](/docs/prow/k3d-no-ssh.md) - This document provides simple instructions, with examples, on how to prepare a ProwJob to use a K3d cluster and Docker.

[Label_sync](/docs/prow/label_sync.md) - Label_sync updates or migrates GitHub labels on repositories in a GitHub organisation based on a YAML file. It is triggered as a `ci-prow-label-sync` Prow job.

[Manage component jobs with templates](/docs/prow/manage-component-jobs-with-templates.md) - This document describes how to define, modify, and remove Prow jobs for Kyma components using predefined templates that create both presubmit and postsubmit jobs for your component. Also, this document gives you the steps required to prepare your component for the Prow CI pipeline.

[Obligatory security measures](/docs/prow/obligatory-security-measures.md) - Read about the obligatory security measures to take on a regular basis and when a Kyma organization member leaves the project.

[Presets](/docs/prow/presets.md) - This document contains the list of all Presets available in the [`config.yaml`](/prow/config.yaml) file. Use them to define Prow Jobs for your components.

[Prow Architecture](/docs/prow/prow-architecture.md) - The document outlines Prow architecture and interconnections between different systems and components that are involved in it.

[Prow cluster update](/docs/prow/prow-cluster-update.md) - Updating a Prow cluster requires an improved Prow version. The Kubernetes Prow instance gets updated via a shell script. The shell script offers only a short list of the last pushed container tags and as a result, limits the versions to choose from. To cherry-pick updates, monitor [Prow announcements](https://github.com/kubernetes/test-infra/blob/master/prow/ANNOUNCEMENTS.md) to see when fixes or important changes are merged into the Kubernetes repository. This document describes how to update a Prow cluster using a cherry-picked Prow version.

[HTML lens](/docs/prow/prow-html-lens.md) - Spyglass HTML lens allows to render HTML files in the job results.

[Image autobump ](/docs/prow/prow-jobs-autobump.md) - This document provides an overview of autobump Prow Jobs. 

[Prow Jobs QuickStart](/docs/prow/prow-jobs-quick-start.md) - This document provides an overview of how to quickly start working with Prow jobs.

[Prow Cluster Monitoring Setup](/docs/prow/prow-monitoring.md) - This document describes how to install and manage Prow cluster monitoring that is available at `https://monitoring.build.kyma-project.io`. 

[Quality metrics](/docs/prow/quality-metrics.md) - This document describes reports that provide an overview of the basic quality measures for the Kyma project.

[Security Leaks Scanner](/docs/prow/security_commit_scanner.md) - Security Leaks Scanner is a tool that scans a repository for potential security leaks, thus providing protection against any potential security threats and vulnerabilities. It operates using [Gitleaks](https://github.com/zricethezav/gitleaks), which ensures a thorough and efficient examination of your repository. 

[Prow Test Clusters](/docs/prow/test-clusters.md) - This document gathers information about test clusters that Prow jobs build. All test clusters are built in the `sap-kyma-prow-workloads` project.

[Tide introduction](/docs/prow/tide-introduction-notes.md) - Along with the Prow upgrade, we want to introduce Tide for merging the PRs automatically.

[Prow Workload Clusters](/docs/prow/workload-clusters.md) - This document describes workload clusters on which Prow schedules Pods to execute the logic of a given Prow job. All workload clusters are aggregated under the `kyma-prow` GCP project. We use two workload clusters for trusted and untrusted Prow jobs.

[Changelog Generator](/experimental/changelog-generator/README.md) - This project is a Docker image that is used to generate a changelog in the `kyma` repository. It uses GitHub API to get pull requests with specified labels.

[Prow runtime images](/images/README.md) - This directory contains images that can be used as runtime images for all ProwJobs in Kyma's Prow Instance.

[E2E DinD K3d](/images/e2e-dind-k3d/README.md) - This image contains common tools for all jobs/tasks that test Kyma modules in K3d.

[PR Tag Builder](/pkg/tools/prtagbuilder/README.md) - PR Tag Builder is a tool that finds a pull request number for a commit.

[Cluster](/prow/cluster/README.md) - This folder contains files related to the configuration of the Prow production cluster that are used during the cluster provisioning.

[Resources](/prow/cluster/resources/README.md) - This directory contains Helm charts used by a Prow cluster.

[External Secrets](/prow/cluster/resources/external-secrets/README.md) - Kubernetes Secrets are synchronized with GCP Secret Manager using [External Secrets Operator](https://github.com/external-secrets/external-secrets).

[Monitoring](/prow/cluster/resources/monitoring/README.md) - This chart contains the monitoring stack for a Prow cluster. It also includes custom-defined Grafana dashboards.

[Probot Stale](/prow/cluster/resources/probot-stale/README.md) - This chart contains the `probot-stale` configuration.

[Images](/prow/images/README.md) - > DEPRECATED: Use the [`images`](/images) directory instead.

[Golangci-lint image](/prow/images/golangci-lint/README.md) - This folder contains the Golangci-lint image that is based on the upstream Golangci-lint image. Use it to lint Go source files.

[Probot Stale](/prow/images/probot-stale/README.md) - This folder contains the image for `probot-stale`.

[Vulnerability Scanner](/prow/images/whitesource-scanner/README.md) - This folder contains the WhiteSource Unified Agent image that is based on the Java Buildpack image. Use it to perform WhiteSource vulnerability scans.

[Templates](/templates/README.md) - Jobs and Prow configuration are generated from templates by the Render Templates tool. Check

