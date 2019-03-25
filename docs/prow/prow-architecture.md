# Prow Architecture

The document outlines Prow architecture and interconnections between different systems and components that are involved in it.

The Prow cluster is a Kubernetes instance managed by Google Kubernetes Engine (GKE) as part of the Google Cloud Platform (GCP) project called `kyma-project`.

See an overview of the Prow production cluster, its components, and interactions with GitHub repositories and GCP.  

![Prow architecture overview](./assets/prow-architecture.svg)

## Provisioning and Secrets
The [set-up-workload-cluster.sh](../../prow/set-up-workload-cluster.sh) script provisions the cluster where Prow runs its workload Pods. To enable the Prow cluster to schedule jobs in the workload cluster, you must create a Secret using [create-secrets-for-workload-cluster.sh](../../prow/create-secrets-for-workload-cluster.sh) script. The Prow main cluster gets provisioned by the [install-prow.sh](../../prow/install-prow.sh) script that the Prow administrator executes manually. The script reads the full configuration from the `test-infra` repository. Based on that, the administrator can recreate the clusters at any time. That is why, when you make any changes to the cluster configuration, apply them directly on the repository instead of the runtime. Only the Prow administration team needs access to the runtime. All others should contribute changes through the repository with the review process in place. The administrators upload any new configuration to the cluster.

Secrets are stored in Google Cloud Storage (GCS) in a dedicated bucket and are encrypted by Key Management Service (KMS). At the time of provisioning, the provisioning script reads all Secrets from GCS and installs them as Kubernetes Secrets on the cluster. The script uses a dedicated service account to access cloud storage. This account is not present as a Secret at runtime.

> **NOTE:** For more information about Secret management, read the [Prow Secret Management](./prow-secrets-management.md) document.

## Components
Prow components access the RBAC-protected API server using dedicated service accounts and are communicating without having TLS enabled.

### Deck
Deck is exposed through an Ingress definition which has TLS enabled using a certificate issued for `status.build.kyma-project.io`. Deck serves a UI that you can access as an anonymous user to view build statuses. Deck can only view and list the jobs and the job logs.

### Hook
Hook is exposed through the same Ingress as Deck using a different path which is `https://status.build.kyma-project.io/hook`. It listens for GitHub events triggered by the external GitHub system. The external address to the Hook component gets configured in GitHub as a webhook using a token as a Secret. That token gets generated during the provisioning process and is configured for the Hook component. Hook calls the installed plugins on receiving a GitHub event.

### Plank
Plank checks regularly if there are new ProwJob resources, executes the related job, and applies the Pod specification to the cluster. A ProwJob gets created usually by the Trigger plugin based on an event from GitHub, or periodically by the Horologium component.

### Horologium
Horologium triggers periodic jobs from the `job` folder based on a predefined trigger period.

### Sinker
Sinker scans for jobs older than one day and cleans up their Pods.

### Branch Protector
Branch Protector is a Prow component that is responsible for defining branch protection settings on GitHub repositories. It updates protection settings on GitHub repositories every 30 minutes. It takes configuration from the `config.yaml` file on the cluster.

### Prow Addons Controller Manager

The Prow Addons Controller Manager contains all custom controller extensions for Prow infrastructure, such as the Notifier controller. Notifier watches all ProwJobs and sends notifications to a given Slack channel. Find detailed documentation [here](../../development/prow-addons-ctrl-manager/README.md).

>**NOTE:** Unlike other Prow components, the Prow Addons Controller Manager is a tool developed in Kyma.

## Plugins
There are different kinds of plugins that react to GitHub events forwarded by the Hook component. Plugins are configured per repository using `plugins.yaml`.
Prow plugins applied for the Kyma project include:
- **trigger** that matches the received event against the job configuration from the `jobs` folder. If it finds the match, it creates a new ProwJob resource.
- **cat** that checks if there is a new GitHub event for a `/meow` comment on a PR. If it finds it, it adds a cat image to the related PR. For that purpose, it uses the GitHub token available as a Kubernetes Secret.
- **config-updater** that reads the configuration from `config.yaml`, `plugins.yaml`, and the `jobs` folder, and updates it on the production cluster after the merge to the `master` branch. This plugin is only configured for the `test-infra` repository.

## ProwJobs
Different build jobs are specified in the `jobs` folder per repository. Each of them uses different kind of trigger conditions. Depending on the trigger, a component becomes active to create a Prow-specific ProwJob resource that represents a given job execution. At a later time, a real Pod gets created by the Plank based on the Pod specification provided in the `jobs` folder. Inside the Pod, a container executes the actual build logic. When the process is finished, the Sinker component cleans up the Pod.

> **NOTE:** A job cannot access the K8s API.

## Dynamic provisioning using GKE or Google Compute Engine (GCE)
The integration job performs integration tests against real clusters. To achieve this, it creates and deletes either the managed Kubernetes clusters using GKE or Virtual Machines (VM) with Minikube installed on them. The integration job uses the Secret configured for a dedicated Google service account.

## Publish images to Google Container Registry (GCR)
Every job can have a Secret configured to upload Docker images to GCR. That Secret belongs to a dedicated Google service account.
Prow in Kyma uses the Docker-in-Docker (dind) approach to build a Docker image as part of a job.

## Build logs on GCS
Build logs are archived by Plank on GCS in a dedicated bucket. The bucket is configured to have a Secret with a dedicated Google service account for GCS.

## Generate development artifacts

There are two jobs that generate artifacts which allow you to install Kyma on a cluster either from the `master` branch or from a pull request changes:
- `pre-master-kyma-development-artifacts`
- `post-master-kyma-development-artifacts`

>**NOTE:** For pull requests, the job is executed only if the introduced changes have an impact on the installed Kyma version.

Find the jobs definitions in [this](https://github.com/kyma-project/test-infra/blob/master/prow/jobs/kyma/kyma-development-artifacts.yaml) file.

All artifacts are stored in the publicly available bucket under the `gs://kyma-development-artifacts/` location. The bucket has a defined lifecycle management rule to automatically delete files older than 60 days. These are the exact artifacts locations:
* For pull requests: `gs://kyma-development-artifacts/PR-<number>`
* For changes to the `master` branch: `gs://kyma-development-artifacts/master-<commit_sha>`
* For the latest changes in the master branch:  `gs://kyma-development-artifacts/master`

A directory with artifacts consists of the following files:
- `kyma-installer-cluster.yaml` to deploy Kyma installer
- `kyma-config-cluster.yaml` to configure Kyma installation
- `is-installed.sh` to verify if Kyma installation process is finished
