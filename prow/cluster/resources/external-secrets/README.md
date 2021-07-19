# External secrets

## Overview

Kubernetes secrets are synchronized with GCP Secret Manager using [Kubernetes External Secrets](https://github.com/external-secrets/kubernetes-external-secrets).

# Installation

To install `kubernetes-external-secrets` on a untrusted cluster in `external-secrets` namespace:

1. `kubectl create namespace external-secrets`
2. Add `sa-secret-manager-untrusted` secret containing credentials for GCP service account with permissions to access secrets.
3. `helm install -f prow/cluster/resources/external-secrets/values_untrusted.yaml -n external-secrets kubernetes-external-secrets external-secrets/kubernetes-external-secrets`
# Configuration

Secrets can be stored as text in GCP Secret Manager and be mapped to a Kubernetes secret with one key. 

See an example:

```yaml
apiVersion: kubernetes-client.io/v1
kind: ExternalSecret
metadata:
  name: plainSecret # name of the k8s external secret and the k8s secret
spec:
  backendType: gcpSecretsManager
  projectId: my-gcp-project
  data:
    - key: gcp-plain-secret # name of the GCP secret
      name: token # key name in the k8s secret
      version: latest # version of the GCP secret
```

Secrets can also be stored as JSON in GCP Secret Manager and be mapped to a Kubernetes secret with multiple keys. 

See an example:

```yaml
apiVersion: kubernetes-client.io/v1
kind: ExternalSecret
metadata:
  name: secretName # name of the k8s external secret and the k8s secret
spec:
  backendType: gcpSecretsManager
  projectId: my-gcp-project
  data:
    - key: gcp-json-secret # name of the GCP secret
      name: keyName # key name in the k8s secret
      version: latest # version of the GCP secret
      property: keyName # name of the field in the GCP secret JSON, unused for plain values
    - key: gcp-json-secret # name of the GCP secret
      name: anotherKey # key name in the k8s secret
      version: latest # version of the GCP secret
      property: anotherKey # name of the field in the GCP secret JSON, unused for plain values
```
