# Authorization

## Required GCP Permissions

To deploy a Prow cluster, configure the following service accounts in the GCP project you own.

| Service account name          | Usage                                                      | Required roles |
| :---------------------------- | :----------------------------------------------------------| :------------- |
| **sa-gcs-plank**              | Used by Prow plan microservice | `Storage Object Admin`
| **sa-gke-kyma-integration**   | Running integration tests on GKE cluster | `Compute Admin`, `Kubernetes Engine Admin`, `Kubernetes Engine Cluster Admin`, `DNS Administrator`, `Service Account User`, `Storage Admin`
| **sa-kyma-artifacts**         | Saving release artifacts to the GCS bucket | `Storage Object Admin`
| **sa-vm-kyma-integration**    | Running integration tests on minikube | `Compute Instance Admin (beta)`, `Compute OS Admin Login`, `Service Account User`
| **sa-gcr-push-kyma-project**  | Publishing docker images | `Storage Admin`

## Kubernetes RBAC rules on Prow cluster

Following cluster roles exist on Prow cluster:
- cert-manager - is able to manage following resources:
    - `certificates.certmanager.k8s.io` 
    - `issuers.certmanager.k8s.io`
    - `clusterissuers.certmanager.k8s.io`
    - `configmaps`
    - `secrets`
    - `events`
    - `services`
    - `pods`
    - `ingresses.extensions`

The `cluster-admin` kubernetes role is granted to `Tiller` service account.  

Following roles exist on Prow cluster:

| Role name   | managed resources | available verbs |
| :---------- | :---------------- | :-------------- |
| **deck** | - `prowjobs.prow.k8s.io`  <br> - `pods/log` | get, list <br> get |
| **horologium** | - `prowjobs.prow.k8s.io`  <br> - `pods` | delete, list <br> delete, list |
| **plank** | - `prowjobs.prow.k8s.io` <br> - `pods` | create, list, update <br> create, list, delete |
| **sinker** | - `prowjobs.prow.k8s.io` <br> - `pods` | delete, list <br> delete, list |
| **hook** | - `prowjobs.prow.k8s.io` <br> - `configmaps` | create, get <br> get, update |
| **tide** | - `prowjobs.prow.k8s.io` |  create, list  |

## User permissions on GitHub

Prow is responsible for starting tests in reaction to certain Github events. For security reasons, the `trigger` plugin ensures that test jobs are run only on pull requests created or verified by trusted users.

### Trusted users
Members of the `kyma-project` organization are considered trusted users. Trigger starts jobs automatically when a trusted user opens a pull request or commits changes to a pull request branch. Alternatively, trusted collaborators can start jobs manually via the `/test all`, `/test {JOB_NAME}` and `/retest` commands, even if a particular pull request was created by an external user. 

### External contributors
External contributors are users outside the `kyma-project` organization. Trigger does not automatically start test jobs on pull requests created by external contributors. Furthermore, external contributors are not allowed to manually run tests on their own pull requests.

> **NOTE:** External contributors can still trigger tests on pull requests created by trusted users.

## Authorization decisions enforced by Prow

Action on Prow can be only triggered by webhooks. To configure them you need to provide two secrets:
- hmac-token - used to validate webhook
- oauth-token - GitHub bot access token

For more details see [Prow documentation](https://github.com/kubernetes/test-infra/blob/master/prow/getting_started_deploy.md#create-the-github-secrets).
