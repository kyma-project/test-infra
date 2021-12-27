# External Secrets

## Overview

Kubernetes Secrets are synchronized with GCP Secret Manager using [Kubernetes External Secrets](https://github.com/external-secrets/kubernetes-external-secrets).

## Installation

Follow these steps to install `kubernetes-external-secrets` on an untrusted cluster in the `external-secrets` Namespace.

1. Create the `external-secrets` Namespace. Run:

   ```bash
   kubectl create namespace external-secrets
   ```

2. Add the `sa-secret-manager-untrusted` Secret containing credentials for a GCP service account with permission to access Secrets.

3. Add the `external-secrets` Helm repository. Use the following command:

   ```bash
   helm repo add external-secrets https://external-secrets.github.io/kubernetes-external-secrets/
   ```

4. Install the `external-secrets/kubernetes-external-secrets` Helm chart. Run:

   ```bash
   helm install -f prow/cluster/resources/external-secrets/values_untrusted.yaml -n external-secrets kubernetes-external-secrets external-secrets/kubernetes-external-secrets
   ```

## Configuration

Secrets can be stored as text in GCP Secret Manager and be mapped to a Kubernetes Secret with one key. 

See an example:

```yaml
apiVersion: kubernetes-client.io/v1
kind: ExternalSecret
metadata:
  name: plainSecret # name of the k8s external Secret and the k8s Secret
spec:
  backendType: gcpSecretsManager
  projectId: my-gcp-project
  data:
    - key: gcp-plain-secret # name of the GCP Secret
      name: token # key name in the k8s Secret
      version: latest # version of the GCP Secret
```

Secrets can also be stored as JSON in GCP Secret Manager and be mapped to a Kubernetes Secret with multiple keys. 

See an example:

```yaml
apiVersion: kubernetes-client.io/v1
kind: ExternalSecret
metadata:
  name: secretName # name of the k8s external Secret and the k8s Secret
spec:
  backendType: gcpSecretsManager
  projectId: my-gcp-project
  data:
    - key: gcp-json-secret # name of the GCP Secret
      name: keyName # key name in the k8s Secret
      version: latest # version of the GCP Secret
      property: keyName # name of the field in the GCP Secret JSON, unused for plain values
    - key: gcp-json-secret # name of the GCP Secret
      name: anotherKey # key name in the k8s Secret
      version: latest # version of the GCP Secret
      property: anotherKey # name of the field in the GCP Secret JSON, unused for plain values
```
>**NOTE:** The trusted and untrusted files are only applied to trusted or untrusted clusters respectively. While the workload file is applied to both trusted and untrusted clusters.
   The presubmit and pj-tester jobs are executed on untrusted clusters, while the periodic jobs are run on trusted clusters. Adding a Secret to the proper file allows the user to specify which type of clusters should have access to the Secret.


## External Secrets Checker

External Secrets Checker checks if all External Secrets synchronized successfully, and if all Secrets have corresponding External Secrets.

To install External Secrets Checker run the following command:

`kubectl apply -f external_secrets_checker_prow.yaml`
