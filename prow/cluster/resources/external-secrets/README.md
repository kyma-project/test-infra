# External secrets

## Overview

Kubernetes secrets are synchronized with GCP secret manager using [Kubernetes External Secrets](https://github.com/external-secrets/kubernetes-external-secrets).

# Installation

To install `kubernetes-external-secrets` on a cluster use following command:

```bash
helm install -f values_untrusted.yaml -n external-secrets kubernetes-external-secrets external-secrets/kubernetes-external-secrets
```

# Configuration

See the example secret config which creates `secretName` kubernetes secret using `value` field of `my-gcp-secret-name` secret from the GCP secret manager.
```yaml
apiVersion: kubernetes-client.io/v1
kind: ExternalSecret
metadata:
  name: secretName # name of the k8s external secret and the k8s secret
spec:
  backendType: gcpSecretsManager
  projectId: my-gcp-project
  data:
    - key: my-gcp-secret-name # name of the GCP secret
      name: my-kubernetes-secret-key-name # key name in the k8s secret
      version: latest # version of the GCP secret
      property: value # name of the field in the GCP secret JSON, unused for plain values
```
