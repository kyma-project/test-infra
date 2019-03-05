# Authorization

## Required GCP Permissions

To deploy a Prow cluster, configure the following service accounts in the GCP project you own.

| Service account name          | Usage                                                      | Required roles |
| :---------------------------- | :----------------------------------------------------------| :------------- |
| **sa-gcs-plank**              | Used by Prow plank microservice. | `Storage Object Admin` (`roles/storage.objectAdmin`) in the `kyma-prow-logs` bucket
| **sa-gke-kyma-integration**   | Runs integration tests on a GKE cluster. | `Compute Admin` (`roles/compute.admin`), `Kubernetes Engine Admin` (`roles/container.admin`), `Kubernetes Engine Cluster Admin` (`roles/container.clusterAdmin`), `DNS Administrator` (`roles/dns.admin`), `Service Account User` (`roles/iam.serviceAccountUser`), `Storage Admin` (`roles/storage.admin`)
| **sa-kyma-artifacts**         | Saves release and development artifacts to the GCS bucket. | `Storage Object Admin` (`roles/storage.objectAdmin`)
| **sa-vm-kyma-integration**    | Runs integration tests on Minikube. | `Compute Instance Admin (beta)` (`roles/compute.instanceAdmin`), `Compute OS Admin Login` (`roles/compute.osAdminLogin`), `Service Account User` (`roles/iam.serviceAccountUser`)
| **sa-gcr-push-kyma-project**  | Publishes Docker images. | `Storage Object Creator` (`roles/storage.objectCreator`) in the `eu.artifacts.kyma-project.appspot.com` and `artifacts.kyma-project.appspot.com` buckets


## Kubernetes RBAC roles on Prow cluster

### Cluster Roles

The `cert-manager` Cluster Role is the only Cluster Role required to deploy a Prow cluster. It manages the following resources:

    - `certificates.certmanager.k8s.io` 
    - `issuers.certmanager.k8s.io`
    - `clusterissuers.certmanager.k8s.io`
    - `configmaps`
    - `secrets`
    - `events`
    - `services`
    - `pods`
    - `ingresses.extensions`

> **NOTE:** There is no separate Cluster Role for Tiller. Instead, the `cluster-admin` Kubernetes role is granted to Tiller's service account.

### Roles

Following roles exist on Prow cluster:

| Role name   | Managed resources | Available actions |
| :---------- | :---------------- | :-------------- |
| **deck** | - `prowjobs.prow.k8s.io`  <br> - `pods/log` | get, list <br> get |
| **horologium** | - `prowjobs.prow.k8s.io`  <br> - `pods` | delete, list <br> delete, list |
| **plank** | - `prowjobs.prow.k8s.io` <br> - `pods` | create, list, update <br> create, list, delete |
| **sinker** | - `prowjobs.prow.k8s.io` <br> - `pods` | delete, list <br> delete, list |
| **hook** | - `prowjobs.prow.k8s.io` <br> - `configmaps` | create, get <br> get, update |
| **tide** | - `prowjobs.prow.k8s.io` |  create, list  |

## User permissions on GitHub

Prow starts tests when triggered by certain Github events. For security reasons, the `trigger` plugin ensures that the test jobs are run only on pull requests (PR) created or verified by trusted users.

### Trusted users
All members of the `kyma-project` organization are considered trusted users. The `trigger` plugin starts jobs automatically when a trusted user opens a PR or commits changes to a PR branch. Alternatively, trusted collaborators can start jobs manually through the `/test all`, `/test {JOB_NAME}` and `/retest` commands, even if a particular PR was created by an external user. 

### External contributors
All users that are not members of the `kyma-project` organization are considered external contributors. The `trigger` plugin does not automatically start test jobs on PRs created by external contributors. Furthermore, external contributors are not allowed to manually run tests on their own PRs.

> **NOTE:** External contributors can still trigger tests on PRs created by trusted users.

## Authorization decisions enforced by Prow

Actions on Prow can be triggered only by webhooks. To configure them you must create two Github Secrets on your Prow cluster:
- `hmac-token` - used to validate webhook
- `oauth-token` - stores the access token used by the GitHub bot

Follow the official [Prow documentation](https://github.com/kubernetes/test-infra/blob/master/prow/getting_started_deploy.md#create-the-github-secrets) to learn how to create the Secrets.
