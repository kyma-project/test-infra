# Prow Architecture

The document outlines Prow architecture and interconnections between different systems and components that are involved in it.

The Prow cluster is a Kubernetes instance managed by Google Kubernetes Engine (GKE) as part of the Google Cloud Platform (GCP) project called `kyma-project`.

See an overview of the Prow production cluster, its components, and interactions with GitHub repositories and GCP.

![Prow architecture overview](./assets/prow-architecture.svg)

## Provisioning and Secrets
For information on how to provision your own Prow cluster, read official [Prow docs](https://docs.prow.k8s.io/docs/).

## Components
Prow components access the RBAC-protected API server using dedicated service accounts and are communicating without having TLS enabled.

### Crier
Crier takes care of reporting the status of Prow job to the external services like GitHub and Slack. For more information, read [crier.md](./crier.md).

### Deck
Deck is exposed through an Ingress definition which has TLS enabled using a certificate issued for `status.build.kyma-project.io`. Deck serves a UI that you can access as an anonymous user to view build statuses. Deck can only view and list the jobs and the job logs.

### Hook
Hook is exposed through the same Ingress as Deck using a different path which is `https://status.build.kyma-project.io/hook`. It listens for GitHub events triggered by the external GitHub system. The external address to the Hook component gets configured in GitHub as a webhook using a token as a Secret. That token gets generated during the provisioning process and is configured for the Hook component. Hook calls the installed plugins on receiving a GitHub event.

### Prow-controller-manager
Prow-controller-manager (formerly "Plank") checks regularly if there are new Prow job resources, executes the related job, and applies the Pod specification to the cluster. A Prow job gets created usually by the Trigger plugin based on an event from GitHub, or periodically by the Horologium component.

### Horologium
Horologium triggers periodic jobs from the `job` folder based on a predefined trigger period.

### Sinker
Sinker scans for jobs older than one day and cleans up their Pods.

### Branch Protector
Branch Protector is a Prow component that is responsible for defining branch protection settings on GitHub repositories. It updates protection settings on GitHub repositories every 30 minutes. It takes configuration from the `config.yaml` file on the cluster.

### gcsweb
gcsweb is a lightweight web frontend for GCS which allows you to access the content of the **artifacts** tab in Spyglass without the need to log in. For more information on gcsweb read [this](https://github.com/kubernetes/k8s.io/tree/main/apps/gcsweb) document.

### Tide
Tide is a Prow component that automatically checks the acceptance criteria against opened PRs in the repository. If the given PR passes all the criteria, Tide automatically merges it.

## Plugins
There are different kinds of plugins that react to GitHub events forwarded by the Hook component. Plugins are configured per repository using `plugins.yaml`.
For more information about installed plugins in the `kyma-project` and `kyma-incubator` organisations, refer to the [plugins.yaml](../../prow/plugins.yaml) file.

## Prow Jobs
Different build jobs are specified in the `jobs` folder per repository. Each of them uses different kind of trigger conditions. Depending on the trigger, a component becomes active to create a Prow-specific Prow job resource that represents a given job execution. At a later time, a real Pod gets created by the Plank based on the Pod specification provided in the `jobs` folder. Inside the Pod, a container executes the actual build logic. When the process is finished, the Sinker component cleans up the Pod.

> **NOTE:** A job cannot access the K8s API.

## Dynamic Provisioning Using GKE or Google Compute Engine (GCE)
The integration job performs integration tests against real clusters. To achieve this, it creates and deletes either the managed Kubernetes clusters using GKE or Virtual Machines (VM) with k3d installed on them. The integration job uses the Secret configured for a dedicated Google service account.

## Publish Images to Google Container Registry (GCR)
Every job can have a Secret configured to upload Docker images to GCR. That Secret belongs to a dedicated Google service account.
Prow in Kyma uses the Docker-in-Docker (dind) approach to build a Docker image as part of a job.

## Build Logs on GCS
Build logs are archived by Plank on GCS in a dedicated bucket. The bucket is configured to have a Secret with a dedicated Google service account for GCS.

## Generate Development Artifacts

There are two jobs that generate artifacts which allow you to install Kyma on a cluster either from the `main` branch or from a pull request changes:
- `pre-main-kyma-development-artifacts`
- `post-main-kyma-development-artifacts`

>**NOTE:** For pull requests, the job is executed only if the introduced changes have an impact on the installed Kyma version.

All artifacts are stored in the publicly available bucket under the `gs://kyma-development-artifacts/` location. The bucket has a defined lifecycle management rule to automatically delete files older than 60 days. These are the exact artifacts locations:
* For pull requests: `gs://kyma-development-artifacts/PR-<number>`
* For changes to the `main` branch: `gs://kyma-development-artifacts/master-<commit_sha>`
* For the latest changes in the `main` branch: `gs://kyma-development-artifacts/master`

A directory with artifacts consists of the following files:
- `kyma-installer-cluster.yaml` to deploy Kyma installer
- `is-installed.sh` to verify if Kyma installation process is finished
- `tiller.yaml` to install Tiller
