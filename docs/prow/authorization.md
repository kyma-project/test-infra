# Authorization

## Required GCP Permissions

Every Prow cluster should be deployed in its own GCP project. You need an owner access to deploy Prow and configure it properly. In order to deploy a Prow cluster, configure the following service accounts in the GCP project you own.

| Service account name          | Description                                                      | Required roles                                                                                       |
| :---------------------------- | :--------------------------------------------------------------- | :--------------------------------------------------------------------------------------------------- |
| **sa-gcs-plank**              | Service account used by Prow plan microservice                   | `Storage Object Admin`
| **sa-gke-kyma-integration**   | Service account used to run integration tests on GKE cluster     | `Compute Admin, Kubernetes Engine Admin, Kubernetes Engine Cluster Admin, DNS Administrator, Service Account User, Storage Admin`
| **sa-kyma-artifacts**         | Service account used to save release artifacts to the GCS bucket | `Storage Object Admin`
| **sa-vm-kyma-integration**    | Service account used to run integration tests on minikube        | `Compute Instance Admin (beta), Compute OS Admin Login, Service Account User`
| **sa-gcr-push-kyma-project**  | Service account used to publish docker images                    | `Storage Admin`

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

| Role name      | Description                                                      |
| :------------- | :--------------------------------------------------------------- |
| **deck**       | Role allowing to get, list `prowjobs.prow.k8s.io` resources and to get `pods/log` resources|
| **horologium** | Role allowing to delete, list `prowjobs.prow.k8s.io` resources and to delete, list `pods` resources |
| **plank**      | Role allowing to create, list, update `prowjobs.prow.k8s.io` resources and to create, delete, list` pods` resources |
| **sinker**     | Role allowing to delete, list `prowjobs.prow.k8s.io` resources and to delete, list `pods` resources |
| **hook**       | Role allowing to create, get `prowjobs.prow.k8s.io` resources and to update, get `configmaps` resources |
| **tide**       | Role allowing to create, list `prowjobs.prow.k8s.io` resources | 

## User permissions on GitHub

## Authorization decisions enforced by Prow

Action on Prow can be only triggered by webhooks. To configure them you need to provide two secrets:
- hmac-token - used to validate webhook
- oauth-token - GitHub bot access token

For more details see [Prow documentation](https://github.com/kubernetes/test-infra/blob/master/prow/getting_started_deploy.md#create-the-github-secrets).