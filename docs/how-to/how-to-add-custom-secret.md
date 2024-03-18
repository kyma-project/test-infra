# Add custom secret to Prow

This tutorial shows how to add and use a custom secret in the Prow pipeline.

## Procedure

1. Create a new secret in the Google Secret Manager, please follow [the guide](./how-to-name-secret.md) for naming convention.
2. Apply the necessary permissions. Add the `secret-manager-trusted@sap-kyma-prow.iam.gserviceaccount.com` principal with the `Secret Manager Secret Accessor` role if the secret will be used only for a postsubmit or release job. If you are creating a secret for a presubmit job, use the `secret-manager-untrusted@sap-kyma-prow.iam.gserviceaccount.com` principal with the same role. If you want to use the secret in presubmit and postsubmit jobs, apply both principals.

![permissions](./secret-manager-permissions.png)

3. Update External Secrets Operator YAML file.

    Add External Secret definitions to one of the following files:
    - [external_secrets_trusted.yaml](https://github.com/kyma-project/test-infra/blob/main/prow/cluster/resources/external-secrets/external_secrets_trusted.yaml) if the Secret is applied only on trusted cluster (applicable for postsubmit or release job).
    - [external_secrets_untrusted.yaml](https://github.com/kyma-project/test-infra/blob/main/prow/cluster/resources/external-secrets/external_secrets_untrusted.yaml) if the Secret is applied only on untrusted cluster (applicable for presubmit job).
    - [external_secrets_workloads.yaml](https://github.com/kyma-project/test-infra/blob/main/prow/cluster/resources/external-secrets/external_secrets_workloads.yaml) if the Secret is applied on both clusters (applicable for presubmit and postsubmit jobs).

4. Apply the Secrets manually in the Prow cluster as Kubernetes External Secret.

5. Create ProwJob Preset in [prow-config.yaml ](../../prow/config.yaml) that maps the Secret to the variable or to the file.

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

    Now you can use the Preset in your job definition and refer to the Secret in your pipeline.
