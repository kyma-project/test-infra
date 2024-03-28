# External Secrets

## Overview

Kubernetes Secrets are synchronized with Google Cloud Secret Manager using [External Secrets Operator](https://github.com/external-secrets/external-secrets).

## Installation

Follow these steps to install `external-secrets` in an untrusted cluster in the `external-secrets` namespace.

1. Create the `external-secrets` namespace. Run:

   ```bash
   kubectl create namespace external-secrets
   ```

2. Add the `external-secrets` Helm repository. Use the following command:

   ```bash
   helm repo add external-secrets https://charts.external-secrets.io
   ```

3. Install the `external-secrets/kubernetes-external-secrets` Helm chart. Run:

   ```bash
   helm install -n external-secrets external-secrets external-secrets/external-secrets -f prow/cluster/resources/external-secrets/values_untrusted.yaml
   ```

4. Map the `external-secrets/secret-manager-untrusted` Kubernetes service account to a Google Cloud service account with permission to access Secrets. Run:

  ```bash
  gcloud iam service-accounts add-iam-policy-binding --role roles/iam.workloadIdentityUser --member "serviceAccount:sap-kyma-prow.svc.id.goog[external-secrets/secret-manager-untrusted]" secret-manager-untrusted@sap-kyma-prow.iam.gserviceaccount.com
  ```
5. Create a new Secret Store. Run: 
  ```bash
  kubectl apply -f prow/cluster/resources/external-secrets/secrets_store.yaml
  ```
## Configuration

Secrets can be stored as text in Google Cloud Secret Manager and be mapped to a Kubernetes Secret with one key. 

See an example:

```yaml
apiVersion: kubernetes-client.io/v1
kind: ExternalSecret
metadata:
  name: plainSecret # name of the k8s external Secret and the k8s Secret
spec:
  secretStoreRef:
    name: gcp-secretstore # name of the Secret store
    kind: ClusterSecretStore
  refreshInterval: "10m" # time between secret synchronization
  target:
    deletionPolicy: "Delete" # delete secret when External Secret is deleted
  data:
    - secretKey: token # key name in the k8s Secret
      remoteRef:
        key: gcp-plain-secret # name of the GCP Secret
        version: latest # version of the GCP Secret
```

Secrets can also be stored as JSON in Google Cloud Secret Manager and be mapped to a Kubernetes Secret with multiple keys. 

See an example:

```yaml
apiVersion: kubernetes-client.io/v1
kind: ExternalSecret
metadata:
  name: secretName # name of the k8s external Secret and the k8s Secret
spec:
  secretStoreRef:
    name: gcp-secretstore # name of the Secret store
    kind: ClusterSecretStore
  refreshInterval: "10m" # time between secret synchronization
  target:
    deletionPolicy: "Delete" # delete secret when External Secret is deleted
  data:
    - secretKey: keyName # key name in the k8s Secret
      remoteRef:
        key: gcp-json-secret # name of the GCP Secret
        property: keyName # name of the field in the GCP Secret JSON, unused for plain values
        version: latest # version of the GCP Secret
    - secretKey: anotherKey # key name in the k8s Secret
      remoteRef:
        key: gcp-json-secret # name of the GCP Secret
        property: anotherKey # name of the field in the GCP Secret JSON, unused for plain values
        version: latest # version of the GCP Secret
```
>**NOTE:** The trusted and untrusted files are only applied to trusted or untrusted clusters respectively. While the workload file is applied to both trusted and untrusted clusters.
   The presubmit and pj-tester jobs are executed on untrusted clusters, while the periodic jobs are run on trusted clusters. Adding a Secret to the proper file allows the user to specify which type of clusters should have access to the Secret.


## External Secrets Checker

External Secrets Checker checks if all External Secrets synchronized successfully, and if all Secrets have corresponding External Secrets.

To install External Secrets Checker run the following command:

`kubectl apply -f external_secrets_checker_prow.yaml`
