# Authorization

## Required GCP Permissions

Every Prow cluster should be deployed in its own GCP project. You need an owner access to deploy Prow and configure it properly. In addition you need following service accounts created:

- sa-gcs-plank - service account used by Prow plan microservice with the `Storage Object Admin` role
- sa-gke-kyma-integration - service account used to run integration tests on GKE cluster with `Compute Admin, Kubernetes Engine Admin, Kubernetes Engine Cluster Admin, DNS Administrator, Service Account User, Storage Admin` roles
- sa-kyma-artifacts - service account used to save release artifacts to the GCS bucket with the `Storage Object Admin` role 
- sa-vm-kyma-integration - service account used to run integration tests on minikube with `Compute Instance Admin (beta), Compute OS Admin Login, Service Account User` roles 
- sa-gcr-push-kyma-project- service account used to publish docker images with the `Storage Admin` role 

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
- deck - is able to get, list prowjobs.prow.k8s.io resources and to get pods/log resources
- horologium - is able to create, list prowjobs.prow.k8s.io resources
- plank - is able to create, list, update prowjobs.prow.k8s.io resources and to create, delete, list pods resources
- sinker - is able to delete, list prowjobs.prow.k8s.io resources and to delete, list pods resources
- hook - is able to create, get prowjobs.prow.k8s.io resources and to update, get configmaps resources
- tide - is able to create, list prowjobs.prow.k8s.io resources


## User permissions on GitHub

## Authorization decisions enforced by Prow