[Code of conduct](U) - Each contributor and maintainer of this project agrees to follow the [community Code of Conduct](https://github.com/kyma-project/community/blob/main/CODE_OF_CONDUCT.md) that relies on the CNCF Code of Conduct. Read it to learn about the agreed standards of behavior, shared values that govern our community, and details on how to report any suspected Code of Conduct violations.

[#Overview](U) - To contribute to this project, follow the rules from the general [CONTRIBUTING.md](https://github.com/kyma-project/community/blob/main/CONTRIBUTING.md) document in the `community` repository.

# U[Test Infra](U) - [![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fkyma-project%2Ftest-infra.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fkyma-project%2Ftest-infra?ref=badge_shield)

[Changelog Generator](U) - This project is a Docker image that is used to generate a changelog in the `kyma` repository. It uses GitHub API to get pull requests with specified labels.

[Development](U) - The purpose of the folder is to store tools developed and used in the `test-infra` repository.

[Create Custom Image](U) - The purpose of this document is to define how to create a new Google Compute Engine [custom image](https://cloud.google.com/compute/docs/images) with required dependencies. You can use the new image to provision virtual machine (VM) instances with all dependencies already installed.

[gardener-rotate](U) - The gardener-rotate tool allows you to generate a new access token for the Gardener service accounts and update kubeconfig stored in the Secret Manager.

[gcbuild](U) - This tool serves as an intelligent wrapper for `gcloud builds submit`. It runs remote build jobs on Google infrastructure with setting automated substitutions, that developers can use. It's built to reduce the complexity of building the Docker images.

[Rotate service account secrets using Cloud Function](U) - Cloud Function creates a new key for a GCP service account and updates the required secret data. The function is triggered by a  Pub/Sub message sent by a secret stored in Secret Manager.

[Cleanup of service account secrets using Cloud Function](U) - The Cloud Function deletes old keys for a GCP service account and updates the required secret data for all service account secrets stored in the Secret Manager. The function is triggered by a Cloud Scheduler job.

[Rotate Gardener service account secrets using Cloud Run](U) - The Cloud Run application creates a new key for a GCP service account, updates the required secret data, and deletes old versions of a key. The function is triggered by a Pub/Sub message sent by a secret stored in Secret Manager.

[Rotate KMS secrets using Cloud Run](U) - The Cloud Run application decrypts and encrypts files in a bucket with the latest version of a KMS key, and deletes old versions of a key. The function is triggered by a HTTP POST request sent by a Cloud Scheduler.

[image-builder](U) - This tool serves as an intelligent wrapper for `kaniko-project/executor`. It reduces the complexity of building Docker images and removes the need of using Docker in Docker when building images in K8s infrastructure.

[image-syncer](U) - image-syncer is used to copy container images from one registry to another.

[Image URL Helper](U) - Image URL Helper is a tool that provides the following subcommands:

[JobGuard ](U) - JobGuard is a simple tool that fetches all statuses for GitHub pull requests and waits for some of them to finish.

[GitHub Webhook Gateway](U) - The GitHub Webhook Gateway is written in Golang. It listens for incoming GitHub Webhook events. It validates a Webhook HMAC secret, converts it to a CloudEvents compliant event and forwards it to the Kyma Event Service. It is meant to run within Kyma as a docker container.

[PubSub Gateway](U) - The PubSub Gateway is written in Golang. It pulls messages from PubSub, converts them into a CloudEvents, and forwards them to the Kyma Event Publisher Proxy.

[Overview](U) - The `issueLabeled` Function is executed by Kyma [Serverless](https://kyma-project.io/docs/components/serverless/) when the registered **issuesevent.labeled** event occurs. All **issuesevent.labeled** events for the `internal-incident` or `customer-incident` labels will be processed by this Function.

[prowjobparser](U) - The prowjobparser is a helper tool which parses all Prow Jobs under the provided path, matches them against the provided label filters, and prints matching Prow Job names to the standard output.

[Test Log Collector](U) - The purpose of the Test Log Collector is to gather logs from the testing Pods and to send them to the appropriate Slack channels.

[Octopus Types](U) - Types in that package has been copied from https://github.com/kyma-incubator/octopus/tree/master/pkg/apis/testing/v1alpha1 in order to solve problems with older dependencies in `octopus` project.

[Tools](U) - This project contains Go applications for the `test-infra` repository.

[certbotauthenticator](U) - Certbotauthenticator is a binary called by the certbot when it generates the certificate. The binary is used in during manual DNS challenge authentication. In the manual mode, the certbot passes the domain name and the authentication token as environment variables to the certbotauthenticator to create a TXT record in the domain. This way, the Let's Encrypt system can validate the domain ownership. After the validation completes, the certbotauthenticator is called again to clean the TXT records.

[Clusters Garbage Collector](U) - This command finds and removes orphaned clusters created by the `kyma-gke-integration` job in a Google Cloud Platform (GCP) project.

[Config Uploader](U) - This command uploads Prow plugins, configuration, and jobs to a Prow cluster. Use it for a newly created Prow cluster and to update changes in the configuration on a cluster from a forked repository.

[Disks Garbage Collector](U) - This command finds and removes orphaned disks created by the `kyma-gke-integration` job in a Google Cloud Platform (GCP) project.

[DNS Cleaner](U) - This command finds and removes DNS entries created by the `kyma-gke-long-lasting` job in a Google Cloud Platform (GCP) project.

[IP Address and DNS Record Garbage Collector](U) - This command finds and removes orphaned IP Addresses and related DNS records created by GKE integration jobs in a Google Cloud Platform (GCP) project.

[External Secrets Checker](U) - This command checks external Secrets synchronization status, and if every Secret has a corresponding external Secret.

[GCR cleaner](U) - This command finds and removes old GCR images created by Jobs in the Google Cloud Platform (GCP) project.

[Github issues](U) - This command queries all open Github issues in an organization or repository, and loads that data to a BigQuery table.

[Github Release](U) - This command creates GitHub releases based on artifacts stored in a Google bucket. Each release requires the following set of artifacts:

[GitHub Statistics](U) - `githubstats` fetches statistics for GitHub issues and prints the following JSON object:

[IP cleaner](U) - This command finds and removes orphaned IP addresses created by jobs in the Google Cloud Platform (GCP) project.

[Job Guard](U) - Job Guard was moved [here](/development/jobguard).

[oomfinder](U) - oomfinder is a small tool designed to run in a Pod on each k8s worker node as a privileged container. It will check if Docker or Containerd is used and attach to its socket to listen for oom events. If an oom event occurs, oomfinder will print a message to `os stdout` with the following details:

[Prow Job tester](U) - Prow Job tester is a tool for testing changes to the Prow Jobs' definitions and code running in Prow Jobs. It uses the production Prow instance to run chosen Prow Jobs with changes from pull requests (PRs) without going through multiple cycles of new PRs, reviews, and merges. The whole development can be done within one cycle.

[Render Templates](U) - The Render Templates is a tool that reads the configuration from a [`config.yaml`](/templates/config.yaml) file and [`data`](/templates/data) files to generate output files, such as Prow component jobs. While the `config.yaml` file can hold configuration for an output file, you can place such data within the data files that hold configuration for related output files. Having separate files with grouped data is cleaner and easier to maintain than one huge config file.

[Virtual Machines Garbage Collector](U) - This command finds and removes orphaned virtual machines (VMs) created by Prow jobs in a Google Cloud Platform (GCP) project.

[YAML merge](U) - This command line tool enables merging yaml files into one single file. For the operation to work, the yaml files must follow the same source path. 

[PR Tag Builder](U) - PR Tag Builder is a tool that finds a pull request number for a commit.

[Documentation guidelines](U) - 1. Each repository must contain an automatically updated index page in `docs` directory.

[[#Overview](U) - To contribute to this project, follow the rules from the general [CONTRIBUTING.md](https://github.com/kyma-project/community/blob/main/CONTRIBUTING.md) document in the `community` repository.](U) - [Changelog Generator](U) - This project is a Docker image that is used to generate a changelog in the `kyma` repository. It uses GitHub API to get pull requests with specified labels.

[Docs](U) - The folder contains documents that provide an insight into Prow configuration, development, and testing.

[Authorization](U) - To deploy a Prow cluster, configure the following service accounts in the GCP project you own.

[Crier](U) - Crier reports the Prow Job status changes. For now, it is responsible for Slack notifications as Plank is still reporting the Prow Job statuses to GitHub.

[Manage component jobs with templates](U) - This document describes how to define, modify, and remove Prow jobs for Kyma components using predefined templates that create both presubmit and postsubmit jobs for your component. Also, this document gives you the steps required to prepare your component for the Prow CI pipeline.

[Obligatory security measures](U) - Read about the obligatory security measures to take on a regular basis and when a Kyma organization member leaves the project.

[Presets](U) - This document contains the list of all Presets available in the [`config.yaml`](/prow/config.yaml) file. Use them to define Prow Jobs for your components.

[Production Cluster Configuration](U) - This instruction provides the steps required to deploy a production cluster for Prow.

[Prow Architecture](U) - The document outlines Prow architecture and interconnections between different systems and components that are involved in it.

[Prow cluster update](U) - Updating a Prow cluster requires an improved Prow version. The Kubernetes Prow instance gets updated via a shell script. The shell script offers only a short list of the last pushed container tags and as a result, limits the versions to choose from. To cherry-pick updates, monitor [Prow announcements](https://github.com/kubernetes/test-infra/blob/master/prow/ANNOUNCEMENTS.md) to see when fixes or important changes are merged into the Kubernetes repository. This document describes how to update a Prow cluster using a cherry-picked Prow version.

[HTML lens](U) - Spyglass HTML lens allows to render HTML files in the job results.

[Image autobump ](U) - This document provides an overview of autobump Prow Jobs. 

[Prow Jobs QuickStart](U) - This document provides an overview of how to quickly start working with Prow jobs.

[Prow jobs](U) - This document provides an overview of Prow jobs.  

[TestGrid](U) - [TestGrid](https://testgrid.k8s.io) is an interactive dashboard for viewing tests results in a grid. It parses JUnit reports for generating a grid view from the tests.

[Prow Cluster Monitoring Setup](U) - This document describes how to install and manage Prow cluster monitoring that is available at `https://monitoring.build.kyma-project.io`. 

[Naming conventions](U) - This document describes the naming conventions for the Prow test instance and its resources hosted in Google Cloud Platform.

[Prow Secrets Management](U) - Some jobs require using sensitive data. Encrypt the data using Key Management Service (KMS) and store it in Google Cloud Storage (GCS).

[Prow Secrets](U) - This document lists all types of Secrets used in the `kyma-prow` and `workload-kyma-prow` clusters, where all Prow Jobs are executed.

[Quality metrics](U) - This document describes reports that provide an overview of the basic quality measures for the Kyma project.

[Prow Test Clusters](U) - This document gathers information about test clusters that Prow jobs build. All test clusters are built in the `sap-kyma-prow-workloads` project.

[Tide introduction](U) - Along with the Prow upgrade, we want to introduce Tide for merging the PRs automatically.

[Prow Workload Clusters](U) - This document describes workload clusters on which Prow schedules Pods to execute the logic of a given Prow job. All workload clusters are aggregated under the `kyma-prow` GCP project. We use two workload clusters for trusted and untrusted Prow jobs.

[Prow](U) - Prow is a Kubernetes-developed system that you can use as a Continuous Integration (CI) tool for validating your GitHub repositories and components, managing automatic validation of pull requests (PRs), applying and removing labels, or opening and closing issues.

[Cluster](U) - This folder contains files related to the configuration of the Prow production cluster that are used during the cluster provisioning.

[Resources](U) - This directory contains Helm charts used by a Prow cluster.

[External Secrets](U) - Kubernetes Secrets are synchronized with GCP Secret Manager using [Kubernetes External Secrets](https://github.com/external-secrets/kubernetes-external-secrets).

[Monitoring](U) - This chart contains the monitoring stack for a Prow cluster. It also includes custom-defined Grafana dashboards.

[Probot Stale](U) - This chart contains the `probot-stale` configuration.

[Images](U) - This folder contains a list of images used in Prow jobs.

[Bootstrap Docker Image](U) - This folder contains the Bootstrap image for Prow infrastructure. Use it for a root image for other Prow images and for generic builds.

[Buildpack Golang Docker Image](U) - This folder contains the Buildpack Golang image that is based on the Bootstrap image. Use it to build Golang components.

[Buildpack Node.js Docker Image](U) - This folder contains the Buildpack Node.js image that is based on the Bootstrap image. Use it to build Node.js components.

[Cleaner Docker Image](U) - This image contains the script which performs a cleanup of the service account profile in the `kyma-project` project.

[Gardener-rotate image](U) - This folder contains the gardener-rotate image that is used to automatically update Gardener kubeconfig secrets.

[Golangci-lint image](U) - This folder contains the Golangci-lint image that is based on the upstream Golangci-lint image. Use it to lint Go source files.

[Kyma integration images](U) - This folder contains the image with tools that are necessary to provision Kyma integration clusters.

[Probot Stale](U) - This folder contains the image for `probot-stale`.

[Prow Tools](U) - The directory contains the Dockerfile for the prow tools image with prebuilt tools used in the prow pipelines.

[Prow Tools](U) - The directory contains the Dockerfile for the prow tools image with prebuilt tools used in the prow pipelines.

[Vulnerability Scanner](U) - This folder contains the WhiteSource Unified Agent image that is based on the Java Buildpack image. Use it to perform WhiteSource vulnerability scans.

[Cluster](U) - The folder contains scripts involved in integration tests.

[Cluster Integration Job](U) - The folder contains the source code for the integration job that installs and tests Kyma on a temporary cluster provisioned on Google Kubernetes Engine (GKE).

[Cluster](U) - The folder contains helper scripts with commonly used functions.

[#Overview](U) - The folder contains files that are directly used by Prow pipeline scripts.

[Cluster](U) - This folder contains configuration files for the Prow workload. This configuration is used during cluster provisioning.

[Templates](U) - Jobs and Prow configuration are generated from templates by the Render Templates tool. Check the [Render Templates documentation](../development/tools/cmd/rendertemplates/README.md) for details about usage.

# U