# Changelog Generator
This project is a Docker image that is used to generate a changelog in the `kyma` repository. It uses GitHub API to get pull requests with specified labels.

[/changelog-generator/README.md](/changelog-generator/README.md)

# Development
The purpose of the folder is to store tools developed and used in the `test-infra` repository.

[/development/README.md](/development/README.md)

# Create Custom Image
The purpose of this document is to define how to create a new Google Compute Engine [custom image](https://cloud.google.com/compute/docs/images) with required dependencies. You can use the new image to provision virtual machine (VM) instances with all dependencies already installed.

[/development/custom-image/README.md](/development/custom-image/README.md)

# gardener-rotate
The gardener-rotate tool allows you to generate a new access token for the Gardener service accounts and update kubeconfig stored in the Secret Manager.

[/development/gardener-rotate/README.md](/development/gardener-rotate/README.md)

# gcbuild
This tool serves as an intelligent wrapper for `gcloud builds submit`. It runs remote build jobs on Google infrastructure with setting automated substitutions, that developers can use. It's built to reduce the complexity of building the Docker images.

[/development/gcbuild/README.md](/development/gcbuild/README.md)

# Rotate service account secrets using Cloud Function
Cloud Function creates a new key for a GCP service account and updates the required secret data. The function is triggered by a  Pub/Sub message sent by a secret stored in Secret Manager.

[/development/gcp/cloud-functions/rotateserviceaccount/README.md](/development/gcp/cloud-functions/rotateserviceaccount/README.md)

# Cleanup of service account secrets using Cloud Function
The Cloud Function deletes old keys for a GCP service account and updates the required secret data for all service account secrets stored in the Secret Manager. The function is triggered by a Cloud Scheduler job.

[/development/gcp/cloud-functions/serviceaccountcleaner/README.md](/development/gcp/cloud-functions/serviceaccountcleaner/README.md)

# Rotate Gardener service account secrets using Cloud Run
The Cloud Run application creates a new key for a GCP service account, updates the required secret data, and deletes old versions of a key. The function is triggered by a Pub/Sub message sent by a secret stored in Secret Manager.

[/development/gcp/cloud-run/gardener-sa-rotate/README.md](/development/gcp/cloud-run/gardener-sa-rotate/README.md)

# Rotate KMS secrets using Cloud Run
The Cloud Run application decrypts and encrypts files in a bucket with the latest version of a KMS key, and deletes old versions of a key. The function is triggered by a HTTP POST request sent by a Cloud Scheduler.

[/development/gcp/cloud-run/kms-rotate/README.md](/development/gcp/cloud-run/kms-rotate/README.md)

# image-builder
This tool serves as an intelligent wrapper for `kaniko-project/executor`. It reduces the complexity of building Docker images and removes the need of using Docker in Docker when building images in K8s infrastructure.

[/development/image-builder/README.md](/development/image-builder/README.md)

# image-syncer
image-syncer is used to copy container images from one registry to another.

[/development/image-syncer/README.md](/development/image-syncer/README.md)

# Image URL Helper
Image URL Helper is a tool that provides the following subcommands:

[/development/image-url-helper/README.md](/development/image-url-helper/README.md)

# JobGuard 
JobGuard is a simple tool that fetches all statuses for GitHub pull requests and waits for some of them to finish.

[/development/jobguard/README.md](/development/jobguard/README.md)

# GitHub Webhook Gateway
The GitHub Webhook Gateway is written in Golang. It listens for incoming GitHub Webhook events. It validates a Webhook HMAC secret, converts it to a CloudEvents compliant event and forwards it to the Kyma Event Service. It is meant to run within Kyma as a docker container.

[/development/kyma-github-connector/githubWebhookGateway/README.md](/development/kyma-github-connector/githubWebhookGateway/README.md)

# PubSub Gateway
The PubSub Gateway is written in Golang. It pulls messages from PubSub, converts them into a CloudEvents, and forwards them to the Kyma Event Publisher Proxy.

[/development/kyma-pubsub-connector/pubSubGateway/README.md](/development/kyma-pubsub-connector/pubSubGateway/README.md)

# Overview
The `issueLabeled` Function is executed by Kyma [Serverless](https://kyma-project.io/docs/components/serverless/) when the registered **issuesevent.labeled** event occurs. All **issuesevent.labeled** events for the `internal-incident` or `customer-incident` labels will be processed by this Function.

[/development/kyma-slack-connector/issueLabeledFunction/README.md](/development/kyma-slack-connector/issueLabeledFunction/README.md)

# prowjobparser
The prowjobparser is a helper tool which parses all Prow Jobs under the provided path, matches them against the provided label filters, and prints matching Prow Job names to the standard output.

[/development/prowjobparser/README.md](/development/prowjobparser/README.md)

# Test Log Collector
The purpose of the Test Log Collector is to gather logs from the testing Pods and to send them to the appropriate Slack channels.

[/development/test-log-collector/README.md](/development/test-log-collector/README.md)

# Octopus Types
Types in that package has been copied from https://github.com/kyma-incubator/octopus/tree/master/pkg/apis/testing/v1alpha1 in order to solve problems with older dependencies in `octopus` project.

[/development/test-log-collector/pkg/resources/clustertestsuite/types/README.md](/development/test-log-collector/pkg/resources/clustertestsuite/types/README.md)

# Tools
This project contains Go applications for the `test-infra` repository.

[/development/tools/README.md](/development/tools/README.md)

# certbotauthenticator
Certbotauthenticator is a binary called by the certbot when it generates the certificate. The binary is used in during manual DNS challenge authentication. In the manual mode, the certbot passes the domain name and the authentication token as environment variables to the certbotauthenticator to create a TXT record in the domain. This way, the Let's Encrypt system can validate the domain ownership. After the validation completes, the certbotauthenticator is called again to clean the TXT records.

[/development/tools/cmd/certbotauthenticator/README.md](/development/tools/cmd/certbotauthenticator/README.md)

# Clusters Garbage Collector
This command finds and removes orphaned clusters created by the `kyma-gke-integration` job in a Google Cloud Platform (GCP) project.

[/development/tools/cmd/clusterscollector/README.md](/development/tools/cmd/clusterscollector/README.md)

# Config Uploader
This command uploads Prow plugins, configuration, and jobs to a Prow cluster. Use it for a newly created Prow cluster and to update changes in the configuration on a cluster from a forked repository.

[/development/tools/cmd/configuploader/README.md](/development/tools/cmd/configuploader/README.md)

# Disks Garbage Collector
This command finds and removes orphaned disks created by the `kyma-gke-integration` job in a Google Cloud Platform (GCP) project.

[/development/tools/cmd/diskscollector/README.md](/development/tools/cmd/diskscollector/README.md)

# DNS Cleaner
This command finds and removes DNS entries created by the `kyma-gke-long-lasting` job in a Google Cloud Platform (GCP) project.

[/development/tools/cmd/dnscleaner/README.md](/development/tools/cmd/dnscleaner/README.md)

# IP Address and DNS Record Garbage Collector
This command finds and removes orphaned IP Addresses and related DNS records created by GKE integration jobs in a Google Cloud Platform (GCP) project.

[/development/tools/cmd/dnscollector/README.md](/development/tools/cmd/dnscollector/README.md)

# External Secrets Checker
This command checks external Secrets synchronization status, and if every Secret has a corresponding external Secret.

[/development/tools/cmd/externalsecretschecker/README.md](/development/tools/cmd/externalsecretschecker/README.md)

# GCR cleaner
This command finds and removes old GCR images created by Jobs in the Google Cloud Platform (GCP) project.

[/development/tools/cmd/gcrcleaner/README.md](/development/tools/cmd/gcrcleaner/README.md)

# Github issues
This command queries all open Github issues in an organization or repository, and loads that data to a BigQuery table.

[/development/tools/cmd/githubissues/README.md](/development/tools/cmd/githubissues/README.md)

# Github Release
This command creates GitHub releases based on artifacts stored in a Google bucket. Each release requires the following set of artifacts:

[/development/tools/cmd/githubrelease/README.md](/development/tools/cmd/githubrelease/README.md)

# GitHub Statistics
`githubstats` fetches statistics for GitHub issues and prints the following JSON object:

[/development/tools/cmd/githubstats/README.md](/development/tools/cmd/githubstats/README.md)

# IP cleaner
This command finds and removes orphaned IP addresses created by jobs in the Google Cloud Platform (GCP) project.

[/development/tools/cmd/ipcleaner/README.md](/development/tools/cmd/ipcleaner/README.md)

# Job Guard
Job Guard was moved [here](../../../jobguard).

[/development/tools/cmd/jobguard/README.md](/development/tools/cmd/jobguard/README.md)

# oomfinder
oomfinder is a small tool designed to run in a Pod on each k8s worker node as a privileged container. It will check if Docker or Containerd is used and attach to its socket to listen for oom events. If an oom event occurs, oomfinder will print a message to `os stdout` with the following details:

[/development/tools/cmd/oomfinder/README.md](/development/tools/cmd/oomfinder/README.md)

# Prow Job tester
Prow Job tester is a tool for testing changes to Prow Jobs and scripts in the `test-infra` repository which are under development. It uses the production Prow instance to run chosen Prow Jobs with code from pull requests (PRs) without going through multiple cycles of new PRs, reviews, and merges. The whole development is done within one PR.

[/development/tools/cmd/pjtester/README.md](/development/tools/cmd/pjtester/README.md)

# Render Templates
The Render Templates is a tool that reads the configuration from a [`config.yaml`](../../../../templates/config.yaml) file and [`data`](../../../../templates/data) files to generate output files, such as Prow component jobs. While the `config.yaml` file can hold configuration for an output file, you can place such data within the data files that hold configuration for related output files. Having separate files with grouped data is cleaner and easier to maintain than one huge config file.

[/development/tools/cmd/rendertemplates/README.md](/development/tools/cmd/rendertemplates/README.md)

# Virtual Machines Garbage Collector
This command finds and removes orphaned virtual machines (VMs) created by Prow jobs in a Google Cloud Platform (GCP) project.

[/development/tools/cmd/vmscollector/README.md](/development/tools/cmd/vmscollector/README.md)

# YAML merge
This command line tool enables merging yaml files into one single file. For the operation to work, the yaml files must follow the same source path. 

[/development/tools/cmd/yamlmerge/README.md](/development/tools/cmd/yamlmerge/README.md)

# PR Tag Builder
PR Tag Builder is a tool that finds a pull request number for a commit.

[/development/tools/pkg/prtagbuilder/README.md](/development/tools/pkg/prtagbuilder/README.md)

# Docs
The folder contains documents that provide an insight into Prow configuration, development, and testing.

[/docs/prow/README.md](/docs/prow/README.md)

# Authorization
To deploy a Prow cluster, configure the following service accounts in the GCP project you own.

[/docs/prow/authorization.md](/docs/prow/authorization.md)

# Crier
Crier reports the Prow Job status changes. For now, it is responsible for Slack notifications as Plank is still reporting the Prow Job statuses to GitHub.

[/docs/prow/crier.md](/docs/prow/crier.md)

# Manage component jobs with templates
This document describes how to define, modify, and remove Prow jobs for Kyma components using predefined templates that create both presubmit and postsubmit jobs for your component. Also, this document gives you the steps required to prepare your component for the Prow CI pipeline.

[/docs/prow/manage-component-jobs-with-templates.md](/docs/prow/manage-component-jobs-with-templates.md)

# Obligatory security measures
Read about the obligatory security measures to take on a regular basis and when a Kyma organization member leaves the project.

[/docs/prow/obligatory-security-measures.md](/docs/prow/obligatory-security-measures.md)

# Presets
This document contains the list of all Presets available in the [`config.yaml`](../../prow/config.yaml) file. Use them to define Prow Jobs for your components.

[/docs/prow/presets.md](/docs/prow/presets.md)

# Production Cluster Configuration
This instruction provides the steps required to deploy a production cluster for Prow.

[/docs/prow/production-cluster-configuration.md](/docs/prow/production-cluster-configuration.md)

# Prow Architecture
The document outlines Prow architecture and interconnections between different systems and components that are involved in it.

[/docs/prow/prow-architecture.md](/docs/prow/prow-architecture.md)

# Prow cluster update
Updating a Prow cluster requires an improved Prow version. The Kubernetes Prow instance gets updated via a shell script. The shell script offers only a short list of the last pushed container tags and as a result, limits the versions to choose from. To cherry-pick updates, monitor [Prow announcements](https://github.com/kubernetes/test-infra/blob/master/prow/ANNOUNCEMENTS.md) to see when fixes or important changes are merged into the Kubernetes repository. This document describes how to update a Prow cluster using a cherry-picked Prow version.

[/docs/prow/prow-cluster-update.md](/docs/prow/prow-cluster-update.md)

# HTML lens
Spyglass HTML lens allows to render HTML files in the job results.

[/docs/prow/prow-html-lens.md](/docs/prow/prow-html-lens.md)

# Image autobump 
This document provides an overview of autobump Prow Jobs. 

[/docs/prow/prow-jobs-autobump.md](/docs/prow/prow-jobs-autobump.md)

# Prow Jobs QuickStart
This document provides an overview of how to quickly start working with Prow jobs.

[/docs/prow/prow-jobs-quick-start.md](/docs/prow/prow-jobs-quick-start.md)

# Prow jobs
This document provides an overview of Prow jobs.  

[/docs/prow/prow-jobs.md](/docs/prow/prow-jobs.md)

# TestGrid
[TestGrid](https://testgrid.k8s.io) is an interactive dashboard for viewing tests results in a grid. It parses JUnit reports for generating a grid view from the tests.

[/docs/prow/prow-k8s-testgrid.md](/docs/prow/prow-k8s-testgrid.md)

# Prow Cluster Monitoring Setup
This document describes how to install and manage Prow cluster monitoring that is available at `https://monitoring.build.kyma-project.io`. 

[/docs/prow/prow-monitoring.md](/docs/prow/prow-monitoring.md)

# Naming conventions
This document describes the naming conventions for the Prow test instance and its resources hosted in Google Cloud Platform.

[/docs/prow/prow-naming-convention.md](/docs/prow/prow-naming-convention.md)

# Prow Secrets Management
Some jobs require using sensitive data. Encrypt the data using Key Management Service (KMS) and store it in Google Cloud Storage (GCS).

[/docs/prow/prow-secrets-management.md](/docs/prow/prow-secrets-management.md)

# Prow Secrets
This document lists all types of Secrets used in the `kyma-prow` and `workload-kyma-prow` clusters, where all Prow Jobs are executed.

[/docs/prow/prow-secrets.md](/docs/prow/prow-secrets.md)

# Quality metrics
This document describes reports that provide an overview of the basic quality measures for the Kyma project.

[/docs/prow/quality-metrics.md](/docs/prow/quality-metrics.md)

# Prow Test Clusters
This document gathers information about test clusters that Prow jobs build. All test clusters are built in the `sap-kyma-prow-workloads` project.

[/docs/prow/test-clusters.md](/docs/prow/test-clusters.md)

# Tide introduction
Along with the Prow upgrade, we want to introduce Tide for merging the PRs automatically.

[/docs/prow/tide-introduction-notes.md](/docs/prow/tide-introduction-notes.md)

# Prow Workload Clusters
This document describes workload clusters on which Prow schedules Pods to execute the logic of a given Prow job. All workload clusters are aggregated under the `kyma-prow` GCP project. We use two workload clusters for trusted and untrusted Prow jobs.

[/docs/prow/workload-clusters.md](/docs/prow/workload-clusters.md)

# Prow
Prow is a Kubernetes-developed system that you can use as a Continuous Integration (CI) tool for validating your GitHub repositories and components, managing automatic validation of pull requests (PRs), applying and removing labels, or opening and closing issues.

[/prow/README.md](/prow/README.md)

# Cluster
This folder contains files related to the configuration of the Prow production cluster that are used during the cluster provisioning.

[/prow/cluster/README.md](/prow/cluster/README.md)

# Resources
This directory contains Helm charts used by a Prow cluster.

[/prow/cluster/resources/README.md](/prow/cluster/resources/README.md)

# External Secrets
Kubernetes Secrets are synchronized with GCP Secret Manager using [Kubernetes External Secrets](https://github.com/external-secrets/kubernetes-external-secrets).

[/prow/cluster/resources/external-secrets/README.md](/prow/cluster/resources/external-secrets/README.md)

# Monitoring
This chart contains the monitoring stack for a Prow cluster. It also includes custom-defined Grafana dashboards.

[/prow/cluster/resources/monitoring/README.md](/prow/cluster/resources/monitoring/README.md)

# Probot Stale
This chart contains the `probot-stale` configuration.

[/prow/cluster/resources/probot-stale/README.md](/prow/cluster/resources/probot-stale/README.md)

# Images
This folder contains a list of images used in Prow jobs.

[/prow/images/README.md](/prow/images/README.md)

# Bootstrap Docker Image
This folder contains the Bootstrap image for Prow infrastructure. Use it for a root image for other Prow images and for generic builds.

[/prow/images/bootstrap/README.md](/prow/images/bootstrap/README.md)

# Buildpack Golang Docker Image
This folder contains the Buildpack Golang image that is based on the Bootstrap image. Use it to build Golang components.

[/prow/images/buildpack-golang/README.md](/prow/images/buildpack-golang/README.md)

# Buildpack Node.js Docker Image
This folder contains the Buildpack Node.js image that is based on the Bootstrap image. Use it to build Node.js components.

[/prow/images/buildpack-node/README.md](/prow/images/buildpack-node/README.md)

# Cleaner Docker Image
This image contains the script which performs a cleanup of the service account profile in the `kyma-project` project.

[/prow/images/cleaner/README.md](/prow/images/cleaner/README.md)

# Gardener-rotate image
This folder contains the gardener-rotate image that is used to automatically update Gardener kubeconfig secrets.

[/prow/images/gardener-rotate/README.md](/prow/images/gardener-rotate/README.md)

# Golangci-lint image
This folder contains the Golangci-lint image that is based on the upstream Golangci-lint image. Use it to lint Go source files.

[/prow/images/golangci-lint/README.md](/prow/images/golangci-lint/README.md)

# Kyma integration images
This folder contains the image with tools that are necessary to provision Kyma integration clusters.

[/prow/images/kyma-integration/README.md](/prow/images/kyma-integration/README.md)

# Probot Stale
This folder contains the image for `probot-stale`.

[/prow/images/probot-stale/README.md](/prow/images/probot-stale/README.md)

# Prow Tools
The directory contains the Dockerfile for the prow tools image with prebuilt tools used in the prow pipelines.

[/prow/images/prow-tools/README.md](/prow/images/prow-tools/README.md)

# Prow Tools
The directory contains the Dockerfile for the prow tools image with prebuilt tools used in the prow pipelines.

[/prow/images/test-untrusted-plugin/README.md](/prow/images/test-untrusted-plugin/README.md)

# Vulnerability Scanner
This folder contains the WhiteSource Unified Agent image that is based on the Java Buildpack image. Use it to perform WhiteSource vulnerability scans.

[/prow/images/whitesource-scanner/README.md](/prow/images/whitesource-scanner/README.md)

# Cluster
The folder contains scripts involved in integration tests.

[/prow/scripts/README.md](/prow/scripts/README.md)

# Cluster Integration Job
The folder contains the source code for the integration job that installs and tests Kyma on a temporary cluster provisioned on Google Kubernetes Engine (GKE).

[/prow/scripts/cluster-integration/README.md](/prow/scripts/cluster-integration/README.md)

# Cluster
The folder contains helper scripts with commonly used functions.

[/prow/scripts/lib/README.md](/prow/scripts/lib/README.md)

## Overview
The folder contains files that are directly used by Prow pipeline scripts.

[/prow/scripts/resources/README.md](/prow/scripts/resources/README.md)

# Cluster
This folder contains configuration files for the Prow workload. This configuration is used during cluster provisioning.

[/prow/workload-cluster/README.md](/prow/workload-cluster/README.md)

# Templates
Jobs and Prow configuration are generated from templates by the Render Templates tool. Check the [Render Templates documentation](../development/tools/cmd/rendertemplates/README.md) for details about usage.

[/templates/README.md](/templates/README.md)

# /test-inventory-integration.md
[/test-inventory-integration.md](/test-inventory-integration.md)

