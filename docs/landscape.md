# Prow Landscape
In the following the landscape of the prow setup is documented. It outlines the different systems and components involved and the interconnection between each other.

The "Prow cluster" is a kubernetes instance managed by the google kubernetes engine as part of the project "kyma-project". In that cluster, the prow components are installed.

![Landscape overview](assets/landscape.svg)

# Provisioning and Secrets
The cluster gets provisioned by a script executed manually by a _admin_. The script will read the full configuration from the test-infra repository so that the cluster can be re-created at any time; all configuratin changes should be applied to the repository instead of to the runtime.
As a consequence only the operations team needs access to the runtime, all others should manipulate the repository having review process in place and cause a re-provisioning.

Secrets are stored on a dedicated cloud storage leveraging the key management service. At provisioning time, all secrets are read by the provsioning script from the cloud storage and are installed as kubernetes secrets to the cluster. For the access to the cloud storage a dedicated service account gets used, not present as secret at runtime.

# Internal components
The prow components are accessing the RBAC protected API server via dedicated service accounts and are communicating without having TLS enabled.

## Deck Dashboard / Anonymous Access
The _deck_ component is exposed via an ingress definition having TLS enabled using a certificate issued for _prow.kyma-project.io_. _deck_ is serving a UI accessible by anonymous to view the build status. The _deck_ pod can view/list jobs only.

## Hook / Github integration
The _hook_ component is exposed using the same ingress as _deck_ using a different path. It is listening for Github events triggered by the external Github system; the external URL to the _hook_ component gets configured in Github as webhook using a token as secret. That token gets generated at the provisioning process and configured at the _hook_ component.
The _hook_ component will call the installed plugins on receive of a Github event.

## Plugins
There are different kind of plugins installed reacting on Github events forwarded by the _hook_ component. Plugins are configured per repository and will be triggered accordingly.
The plugins installed currently are:
- *Trigger* plugin will match the received event against the job configuration of ther `config.yaml` and if required will create a new _Job_ resource. Later the _plank_ component will pick that up and will trigger the job execution.
- *cat* plugin will check a received event for a new `/meow` comment on a PR and will then add a cat image to the related PR. For that it will use the Github token available as kubernetes secret.
- The *slack* plugin (not configured yet) will listen on specific github events and notify on a configured slack channel about it using a slack token configured via kubernetes secret.

## Plank
The _plank_ component will regulary check for new _Job_ resources and will apply the related _pod_ specification to the cluster (start job execution). A _job_ resource gets created usally by the _trigger_ plugin or the _horologium_ component.

## Horologium
The _horologium_ component will match jobs configured in `config.yaml` having a period trigger defined and create job resources in case a trigger is firing.

## Sinker
The _sinker_ component will scan for finished jobs and clean the related pods.

# Prow Jobs
in the `config.yaml` different build jobs are specified per repository using diffent kind of trigger conditions. Dependent on the trigger, a component will get active to create at a certain a prow specific _job_ resource representing a specific job execution. At a later time a real _pod_ gets created by the _plank_ compoment followoing the _pod_ specification provided in the `config.yaml`. Inside the _pod_ a container will execute the actual build logic. When the process has finished, the _sinker_ component will clean up the pod.

In general, a job can execute anykind of logic. It cannot access the API Server.

## Dynamic Provisioning / Google Kubernetes Engine / Google Compute Engine
The integration job will perform integration tests against real clusters. For that it will create and delete managed kubernetes clusters or VMs with minikube installed. The secret for a dedicated google service account is configured for that jobs. 

## Release Artifacts / Google Container Registry
Every job can have a secret configured to upload docker images to GCR. That secret belongs to a dedicated google service account.

For being able to build a docker image as part of a job, the docker-in-docker approach is used.

## Build logs / Google Cloud Storage
Build logs will be archived on google cloud storage.
To be clarified how it works
