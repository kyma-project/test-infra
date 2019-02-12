# Authorization

## Required GCP Permissions

Every Prow cluster should be deployed in its own GCP project. You need an owner access to deploy Prow and configure it properly. In addition you need following service accounts created:

- sa-gcs-plank - service account used by Prow plan microservice with the `Storage Object Admin` role
- sa-gke-kyma-integration - service account used to run integration tests on GKE cluster with `Compute Admin, Kubernetes Engine Admin, Kubernetes Engine Cluster Admin, DNS Administrator, Service Account User, Storage Admin` roles
- sa-kyma-artifacts - service account used to save release artifacts to the GCS bucket with the `Storage Object Admin` role 
- sa-vm-kyma-integration - service account used to run integration tests on minikube with `Compute Instance Admin (beta), Compute OS Admin Login, Service Account User` roles 
- sa-gcr-push-kyma-project- service account used to publish docker images with the `Storage Admin` role 

## Kubernetes RBAC rules on Prow cluster

## User permissions on GitHub

## Authorization decisions enforced by Prow