[Changelog Generator](/changelog-generator/README.md) - This project is a Docker image that is used to generate a changelog in the `kyma` repository. It uses GitHub API to get pull requests with specified labels.

[Development](/development/README.md) - The purpose of the folder is to store tools developed and used in the `test-infra` repository.

[Create Custom Image](/development/custom-image/README.md) - The purpose of this document is to define how to create a new Google Compute Engine [custom image](https://cloud.google.com/compute/docs/images) with required dependencies. You can use the new image to provision virtual machine (VM) instances with all dependencies already installed.

[gardener-rotate](/development/gardener-rotate/README.md) - The gardener-rotate tool allows you to generate a new access token for the Gardener service accounts and update kubeconfig stored in the Secret Manager.

[gcbuild](/development/gcbuild/README.md) - This tool serves as an intelligent wrapper for `gcloud builds submit`. It runs remote build jobs on Google infrastructure with setting automated substitutions, that developers can use. It's built to reduce the complexity of building the Docker images.

[Rotate service account secrets using Cloud Function](/development/gcp/cloud-functions/rotateserviceaccount/README.md) - Cloud Function creates a new key for a GCP service account and updates the required secret data. The function is triggered by a  Pub/Sub message sent by a secret stored in Secret Manager.

[Cleanup of service account secrets using Cloud Function](/development/gcp/cloud-functions/serviceaccountcleaner/README.md) - The Cloud Function deletes old keys for a GCP service account and updates the required secret data for all service account secrets stored in the Secret Manager. The function is triggered by a Cloud Scheduler job.

[Rotate Gardener service account secrets using Cloud Run](/development/gcp/cloud-run/gardener-sa-rotate/README.md) - The Cloud Run application creates a new key for a GCP service account, updates the required secret data, and deletes old versions of a key. The function is triggered by a Pub/Sub message sent by a secret stored in Secret Manager.

[Rotate KMS secrets using Cloud Run](/development/gcp/cloud-run/kms-rotate/README.md) - The Cloud Run application decrypts and encrypts files in a bucket with the latest version of a KMS key, and deletes old versions of a key. The function is triggered by a HTTP POST request sent by a Cloud Scheduler.

[image-builder](/development/image-builder/README.md) - This tool serves as an intelligent wrapper for `kaniko-project/executor`. It reduces the complexity of building Docker images and removes the need of using Docker in Docker when building images in K8s infrastructure.

[image-syncer](/development/image-syncer/README.md) - image-syncer is used to copy container images from one registry to another.

[Image URL Helper](/development/image-url-helper/README.md) - Image URL Helper is a tool that provides the following subcommands:

[JobGuard ](/development/jobguard/README.md) - JobGuard is a simple tool that fetches all statuses for GitHub pull requests and waits for some of them to finish.

[GitHub Webhook Gateway](/development/kyma-github-connector/githubWebhookGateway/README.md) - The GitHub Webhook Gateway is written in Golang. It listens for incoming GitHub Webhook events. It validates a Webhook HMAC secret, converts it to a CloudEvents compliant event and forwards it to the Kyma Event Service. It is meant to run within Kyma as a docker container.

[PubSub Gateway](/development/kyma-pubsub-connector/pubSubGateway/README.md) - The PubSub Gateway is written in Golang. It pulls messages from PubSub, converts them into a CloudEvents, and forwards them to the Kyma Event Publisher Proxy.

[Overview](/development/kyma-slack-connector/issueLabeledFunction/README.md) - The `issueLabeled` Function is executed by Kyma [Serverless](https://kyma-project.io/docs/components/serverless/) when the registered **issuesevent.labeled** event occurs. All **issuesevent.labeled** events for the `internal-incident` or `customer-incident` labels will be processed by this Function.

[prowjobparser](/development/prowjobparser/README.md) - The prowjobparser is a helper tool which parses all Prow Jobs under the provided path, matches them against the provided label filters, and prints matching Prow Job names to the standard output.

[Rotate service account secrets](/development/secrets-rotator/cloud-run/rotate-service-account/README.md) - RotateServiceAccount creates a new key for a GCP service account and updates the required secret data. It's triggered by a  Pub/Sub message sent by a secret stored in Secret Manager. It runs as a cloud run container.

[Test Log Collector](/development/test-log-collector/README.md) - The purpose of the Test Log Collector is to gather logs from the testing Pods and to send them to the appropriate Slack channels.

[Octopus Types](/development/test-log-collector/pkg/resources/clustertestsuite/types/README.md) - Types in that package has been copied from https://github.com/kyma-incubator/octopus/tree/master/pkg/apis/testing/v1alpha1 in order to solve problems with older dependencies in `octopus` project.

[Tools](/development/tools/README.md) - This project contains Go applications for the `test-infra` repository.

[certbotauthenticator](/development/tools/cmd/certbotauthenticator/README.md) - Certbotauthenticator is a binary called by the certbot when it generates the certificate. The binary is used in during manual DNS challenge authentication. In the manual mode, the certbot passes the domain name and the authentication token as environment variables to the certbotauthenticator to create a TXT record in the domain. This way, the Let's Encrypt system can validate the domain ownership. After the validation completes, the certbotauthenticator is called again to clean the TXT records.

[Clusters Garbage Collector](/development/tools/cmd/clusterscollector/README.md) - This command finds and removes orphaned clusters created by the `kyma-gke-integration` job in a Google Cloud Platform (GCP) project.

[Config Uploader](/development/tools/cmd/configuploader/README.md) - This command uploads Prow plugins, configuration, and jobs to a Prow cluster. Use it for a newly created Prow cluster and to update changes in the configuration on a cluster from a forked repository.

[Disks Garbage Collector](/development/tools/cmd/diskscollector/README.md) - This command finds and removes orphaned disks created by the `kyma-gke-integration` job in a Google Cloud Platform (GCP) project.

[DNS Cleaner](/development/tools/cmd/dnscleaner/README.md) - This command finds and removes DNS entries created by the `kyma-gke-long-lasting` job in a Google Cloud Platform (GCP) project.

[IP Address and DNS Record Garbage Collector](/development/tools/cmd/dnscollector/README.md) - This command finds and removes orphaned IP Addresses and related DNS records created by GKE integration jobs in a Google Cloud Platform (GCP) project.

[External Secrets Checker](/development/tools/cmd/externalsecretschecker/README.md) - This command checks external Secrets synchronization status, and if every Secret has a corresponding external Secret.

[GCR cleaner](/development/tools/cmd/gcrcleaner/README.md) - This command finds and removes old GCR images created by Jobs in the Google Cloud Platform (GCP) project.

[Github issues](/development/tools/cmd/githubissues/README.md) - This command queries all open Github issues in an organization or repository, and loads that data to a BigQuery table.

[Github Release](/development/tools/cmd/githubrelease/README.md) - This command creates GitHub releases based on artifacts stored in a Google bucket. Each release requires the following set of artifacts:

[GitHub Statistics](/development/tools/cmd/githubstats/README.md) - `githubstats` fetches statistics for GitHub issues and prints the following JSON object:

[IP cleaner](/development/tools/cmd/ipcleaner/README.md) - This command finds and removes orphaned IP addresses created by jobs in the Google Cloud Platform (GCP) project.

[Job Guard](/development/tools/cmd/jobguard/README.md) - Job Guard was moved [here](/development/jobguard).

[oomfinder](/development/tools/cmd/oomfinder/README.md) - oomfinder is a small tool designed to run in a Pod on each k8s worker node as a privileged container. It will check if Docker or Containerd is used and attach to its socket to listen for oom events. If an oom event occurs, oomfinder will print a message to `os stdout` with the following details:

[Prow Job tester](/development/tools/cmd/pjtester/README.md) - Prow Job tester is a tool for testing changes to the Prow Jobs' definitions and code running in Prow Jobs. It uses the production Prow instance to run chosen Prow Jobs with changes from pull requests (PRs) without going through multiple cycles of new PRs, reviews, and merges. The whole development can be done within one cycle.

[Render Templates](/development/tools/cmd/rendertemplates/README.md) - The Render Templates is a tool that reads the configuration from a [`config.yaml`](/templates/config.yaml) file and [`data`](/templates/data) files to generate output files, such as Prow component jobs. While the `config.yaml` file can hold configuration for an output file, you can place such data within the data files that hold configuration for related output files. Having separate files with grouped data is cleaner and easier to maintain than one huge config file.

[Virtual Machines Garbage Collector](/development/tools/cmd/vmscollector/README.md) - This command finds and removes orphaned virtual machines (VMs) created by Prow jobs in a Google Cloud Platform (GCP) project.

[YAML merge](/development/tools/cmd/yamlmerge/README.md) - This command line tool enables merging yaml files into one single file. For the operation to work, the yaml files must follow the same source path. 

[PR Tag Builder](/development/tools/pkg/prtagbuilder/README.md) - PR Tag Builder is a tool that finds a pull request number for a commit.

[Documentation guidelines](/docs/documentation_guidelines.md) - 1. Each repository must contain an automatically updated index page in `docs` directory.

[Add custom secret to Prow](/docs/how-to/how-to-add-custom-secret.md) - This tutorial shows how to add a custom secret and use it in the Prow pipeline.

[Use Tekton Pipelines with Prow](/docs/how-to/how-to_use_tekton-pipelines-in-prow.md) - Kyma Prow instance supports defining and using Tekton Pipelines as a workload. This gives the developer the ability to use

[Docs](/docs/prow/README.md) - The folder contains documents that provide an insight into Prow configuration, development, and testing.

[Authorization](/docs/prow/authorization.md) - To deploy a Prow cluster, configure the following service accounts in the GCP project you own.

[Crier](/docs/prow/crier.md) - Crier reports the Prow Job status changes. For now, it is responsible for Slack notifications as Plank is still reporting the Prow Job statuses to GitHub.

[Label_sync](/docs/prow/label_sync.md) - Label_sync updates or migrates GitHub labels on repositories in a GitHub organisation based on a YAML file. It is triggered as a `ci-prow-label-sync` Prow job.

[Manage component jobs with templates](/docs/prow/manage-component-jobs-with-templates.md) - This document describes how to define, modify, and remove Prow jobs for Kyma components using predefined templates that create both presubmit and postsubmit jobs for your component. Also, this document gives you the steps required to prepare your component for the Prow CI pipeline.

[Obligatory security measures](/docs/prow/obligatory-security-measures.md) - Read about the obligatory security measures to take on a regular basis and when a Kyma organization member leaves the project.

[Presets](/docs/prow/presets.md) - This document contains the list of all Presets available in the [`config.yaml`](/prow/config.yaml) file. Use them to define Prow Jobs for your components.

[Production Cluster Configuration](/docs/prow/production-cluster-configuration.md) - This instruction provides the steps required to deploy a production cluster for Prow.

[Prow Architecture](/docs/prow/prow-architecture.md) - The document outlines Prow architecture and interconnections between different systems and components that are involved in it.

[Prow cluster update](/docs/prow/prow-cluster-update.md) - Updating a Prow cluster requires an improved Prow version. The Kubernetes Prow instance gets updated via a shell script. The shell script offers only a short list of the last pushed container tags and as a result, limits the versions to choose from. To cherry-pick updates, monitor [Prow announcements](https://github.com/kubernetes/test-infra/blob/master/prow/ANNOUNCEMENTS.md) to see when fixes or important changes are merged into the Kubernetes repository. This document describes how to update a Prow cluster using a cherry-picked Prow version.

[HTML lens](/docs/prow/prow-html-lens.md) - Spyglass HTML lens allows to render HTML files in the job results.

[Image autobump ](/docs/prow/prow-jobs-autobump.md) - This document provides an overview of autobump Prow Jobs. 

[Prow Jobs QuickStart](/docs/prow/prow-jobs-quick-start.md) - This document provides an overview of how to quickly start working with Prow jobs.

[Prow jobs](/docs/prow/prow-jobs.md) - This document provides an overview of Prow jobs.  

[TestGrid](/docs/prow/prow-k8s-testgrid.md) - [TestGrid](https://testgrid.k8s.io) is an interactive dashboard for viewing tests results in a grid. It parses JUnit reports for generating a grid view from the tests.

[Prow Cluster Monitoring Setup](/docs/prow/prow-monitoring.md) - This document describes how to install and manage Prow cluster monitoring that is available at `https://monitoring.build.kyma-project.io`. 

[Naming conventions](/docs/prow/prow-naming-convention.md) - This document describes the naming conventions for the Prow test instance and its resources hosted in Google Cloud Platform.

[KMS Secrets Management](/docs/prow/prow-secrets-management.md) - Some jobs require using sensitive data. Encrypt the data using Key Management Service (KMS) and store it in Google Cloud Storage (GCS).

[Prow Secrets](/docs/prow/prow-secrets.md) - This document lists all types of Secrets used in the `kyma-prow` and `workload-kyma-prow` clusters, where all Prow Jobs are executed.

[Quality metrics](/docs/prow/quality-metrics.md) - This document describes reports that provide an overview of the basic quality measures for the Kyma project.

[Security Leaks Scanner](/docs/prow/security_commit_scanner.md) - Security Leaks Scanner is a tool that scans a repository for potential security leaks, thus providing protection against any potential security threats and vulnerabilities. It operates using [Gitleaks](https://github.com/zricethezav/gitleaks), which ensures a thorough and efficient examination of your repository. 

[Prow Test Clusters](/docs/prow/test-clusters.md) - This document gathers information about test clusters that Prow jobs build. All test clusters are built in the `sap-kyma-prow-workloads` project.

[Tide introduction](/docs/prow/tide-introduction-notes.md) - Along with the Prow upgrade, we want to introduce Tide for merging the PRs automatically.

[Prow Workload Clusters](/docs/prow/workload-clusters.md) - This document describes workload clusters on which Prow schedules Pods to execute the logic of a given Prow job. All workload clusters are aggregated under the `kyma-prow` GCP project. We use two workload clusters for trusted and untrusted Prow jobs.

[Prow](/prow/README.md) - Prow is a Kubernetes-developed system that you can use as a Continuous Integration (CI) tool for validating your GitHub repositories and components, managing automatic validation of pull requests (PRs), applying and removing labels, or opening and closing issues.

[Cluster](/prow/cluster/README.md) - This folder contains files related to the configuration of the Prow production cluster that are used during the cluster provisioning.

[Resources](/prow/cluster/resources/README.md) - This directory contains Helm charts used by a Prow cluster.

[External Secrets](/prow/cluster/resources/external-secrets/README.md) - Kubernetes Secrets are synchronized with GCP Secret Manager using [External Secrets Operator](https://github.com/external-secrets/external-secrets).

[Monitoring](/prow/cluster/resources/monitoring/README.md) - This chart contains the monitoring stack for a Prow cluster. It also includes custom-defined Grafana dashboards.

[Probot Stale](/prow/cluster/resources/probot-stale/README.md) - This chart contains the `probot-stale` configuration.

[Images](/prow/images/README.md) - This folder contains a list of images used in Prow jobs.

[Bootstrap Docker Image](/prow/images/bootstrap/README.md) - This folder contains the Bootstrap image for Prow infrastructure. Use it for a root image for other Prow images and for generic builds.

[Buildpack Golang Docker Image](/prow/images/buildpack-golang/README.md) - This folder contains the Buildpack Golang image that is based on the Bootstrap image. Use it to build Golang components.

[Buildpack Node.js Docker Image](/prow/images/buildpack-node/README.md) - This folder contains the Buildpack Node.js image that is based on the Bootstrap image. Use it to build Node.js components.

[Cleaner Docker Image](/prow/images/cleaner/README.md) - This image contains the script which performs a cleanup of the service account profile in the `kyma-project` project.

[Gardener-rotate image](/prow/images/gardener-rotate/README.md) - This folder contains the gardener-rotate image that is used to automatically update Gardener kubeconfig secrets.

[Golangci-lint image](/prow/images/golangci-lint/README.md) - This folder contains the Golangci-lint image that is based on the upstream Golangci-lint image. Use it to lint Go source files.

[Kyma integration images](/prow/images/kyma-integration/README.md) - This folder contains the image with tools that are necessary to provision Kyma integration clusters.

[Probot Stale](/prow/images/probot-stale/README.md) - This folder contains the image for `probot-stale`.

[Prow Tools](/prow/images/prow-tools/README.md) - The directory contains the Dockerfile for the prow tools image with prebuilt tools used in the prow pipelines.

[Prow Tools](/prow/images/test-untrusted-plugin/README.md) - The directory contains the Dockerfile for the prow tools image with prebuilt tools used in the prow pipelines.

[Vulnerability Scanner](/prow/images/whitesource-scanner/README.md) - This folder contains the WhiteSource Unified Agent image that is based on the Java Buildpack image. Use it to perform WhiteSource vulnerability scans.

[Cluster](/prow/scripts/README.md) - The folder contains scripts involved in integration tests.

[Cluster Integration Job](/prow/scripts/cluster-integration/README.md) - The folder contains the source code for the integration job that installs and tests Kyma on a temporary cluster provisioned on Google Kubernetes Engine (GKE).

[Cluster](/prow/scripts/lib/README.md) - The folder contains helper scripts with commonly used functions.

[#Overview](/prow/scripts/resources/README.md) - The folder contains files that are directly used by Prow pipeline scripts.

[Cluster](/prow/workload-cluster/README.md) - This folder contains configuration files for the Prow workload. This configuration is used during cluster provisioning.

[Templates](/templates/README.md) - Jobs and Prow configuration are generated from templates by the Render Templates tool. Check the [Render Templates documentation](../development/tools/cmd/rendertemplates/README.md) for details about usage.

