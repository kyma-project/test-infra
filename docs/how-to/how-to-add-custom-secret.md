# How to add custom secret to Prow

## Overview

This document describes how to add custom secret and use it in the Prow pipeline.

## Necessary steps

1. Add secret to the Google Secret Manager service and apply necessary permissions. Add `secret-manager-trusted@sap-kyma-prow.iam.gserviceaccount.com` principal with `Secret Manager Seccret Accessor` role if the secret will be used only for postsubmit or release job. If you are creating secret for presubmit job, use `secret-manager-untrusted@sap-kyma-prow.iam.gserviceaccount.com` principal with the same role. If you are planning to use the secret in presubmit and postsubmit jobs - apply both principals.

![permissions](./secret-manager-permissions.png)

2. Update External Secrets Operator yaml file.

Add External Secret definitions to one of the following files:
- [external_secrets_trusted.yaml](https://github.com/kyma-project/test-infra/blob/main/prow/cluster/resources/external-secrets/external_secrets_trusted.yaml) if the secret is applied only on trusted cluster (applicable for postsubmit or release job).
- [external_secrets_untrusted.yaml](https://github.com/kyma-project/test-infra/blob/main/prow/cluster/resources/external-secrets/external_secrets_untrusted.yaml) if the secret is applied only on untrusted cluster (applicable for presubmit job).
- [external_secrets_workloads.yaml](https://github.com/kyma-project/test-infra/blob/main/prow/cluster/resources/external-secrets/external_secrets_workloads.yaml) if the secret is applied on both clusters (applicable for presubmit and postsubmit jobs).

3. Apply secrets manually in the Prow cluster as K8s External Secret.

4. Create prowjob preset in [prow-config.yaml ](https://github.com/kyma-project/test-infra/blob/main/templates/templates/prow-config.yaml) that maps the secret to the variable or to the file.

For example:

```yaml
  - labels:
      preset-kyma-btp-manager-bot-github-token: "true"
    env:
      - name: BOT_GITHUB_TOKEN
        valueFrom:
          secretKeyRef:
            name: kyma-btp-manager-bot-github-token
            key: token
```

Now you can use the preset in your job definition and refer to the secret in your pipeline.
