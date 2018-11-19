# Prow Landscape
In this page, the landscape of the Prow setup is documented. It outlines the different systems and components involved and the interconnection between them.

The "Prow cluster" is a Kubernetes instance managed by Google Kubernetes Engine as part of the project, called "kyma-project". In that cluster, the Prow components are installed.

![Landscape overview](assets/landscape.svg)

# Provisioning and Secrets
The cluster gets provisioned by [install-prow.sh](../prow/install-prow.sh) executed manually by an _admin_. The script will read the full configuration from the test-infra repository so that the cluster can be re-created at any time. Therefore, all the configuration changes should be applied to the repository instead of the runtime.
As a consequence, only the operations team needs access to the runtime, all the others should manipulate the repository having review process in place and cause a re-provisioning.

Secrets are stored on a dedicated bucket in Google Cloud Storage leveraging the cloud Key Management Service. At provisioning time, all secrets are read by the provisioning script from the cloud storage and are installed as Kubernetes Secrets to the cluster. For the access to the cloud storage, a dedicated service account is used, not present as secret at runtime. More information about secret management can be found in [Prow Secret Management](prow-secrets-management.md).

# Internal Components
The Prow components access the RBAC protected API server via dedicated service accounts and are communicating without having TLS enabled.

## Deck Dashboard / Anonymous Access
The _deck_ component is exposed via an Ingress definition having TLS enabled using a certificate issued for _prow.kyma-project.io_. _deck_ is serving a UI accessible anonymously to view the build status. The _deck_ Pod can only view/list the jobs and the job logs.

## Hook / GitHub Integration
The _hook_ component is exposed using the same Ingress as _deck_ using a different path. It is listening for GitHub events triggered by the external GitHub system. The external URL to the _hook_ component gets configured in GitHub as webhook using a token as secret. That token gets generated at the provisioning process and configured for the _hook_ component.
The _hook_ component will call the installed plugins on receive of a GitHub event.

## Plank
The _plank_ component will regulary check for new _ProwJob_ resources and will apply the related Pod specification to the cluster (triggering job execution). A _ProwJob_ resource gets created usally by the _trigger_ plugin or the _horologium_ component.

## Horologium
The _horologium_ component will match jobs configured in `config.yaml` having a trigger period defined and create job resources in case a trigger is firing.

## Sinker
The _sinker_ component will scan for the jobs older than one day and clean up their Pods.

## Plugins
There are different kinds of plugins installed reacting on GitHub events forwarded by the _hook_ component. Plugins are configured per repository using `plugins.yaml` and will be triggered accordingly.
The currently installed plugins are:
- *Trigger* plugin will match the received event against the job configuration of their `config.yaml`, and if required it will create a new _ProwJob_ resource. Later the _plank_ component will pick that up and will trigger the job execution.
- *cat* plugin will check a received event for a new `/meow` comment on a PR and will then add a cat image to the related PR. For that, it will use the GitHub token available as a Kubernetes Secret.
- The *slack* plugin (not yet configured) will listen to the specific GitHub events and publish notifications on a configured Slack channel using a Slack token provided via a Kubernetes Secret.

# Prow Jobs
In the `config.yaml` different build jobs are specified per repository using diffent kinds of trigger conditions. Depending on the trigger, a component will get active to create a Prow specific _ProwJob_ resource representing a specific job execution. At a later time, a real Pod gets created by the _plank_ component following the Pod specification provided in the `config.yaml`. Inside the Pod, a container will execute the actual build logic. When the process is finished, the _sinker_ component will clean up the Pod.

In general, a job can execute any kind of logic. It cannot access to the API Server.

## Dynamic Provisioning using Google Kubernetes Engine (GKE) or Google Compute Engine (GCE)
The integration job will perform integration tests against real clusters. To achieve this, it will create and delete either managed Kubernetes clusters using GKE or Virtual Machines (VM) with Minikube installed. The secret for a dedicated Google service account is configured for these jobs.

## Publising Images to Google Container Registry (GCR)
Every job can have a Secret configured to upload Docker images to GCR. That secret belongs to a dedicated Google service account.

For being able to build a Docker image as part of a job, the Docker-in-Docker (dind) approach is used.

## Build Logs on Google Cloud Storage (GCS)
Build logs are archived on Google Cloud Storage by Plank. It is configured to have a Secret with a dedicated Google service account for GCS.  
