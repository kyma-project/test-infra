# How to name a secret

This tutorial describes how to name a secret in the Google Secret Manager.

## Naming convention

1. The secret name should be in the following format: `<tool for which secret was created>-<entity>`. For example: `github-kyma-bot-token`. From the name it should be clear what the secret contains, a GitHub token for the Kyma bot service account.
   The name should not contain information about where the secret is used. For example: `security-dashboard-github-kyma-bot-token` contains information about the Security Dashboard, which is not necessary. The same secret can be used in different places, so the name should be generic.
   The secret should contain the following fields:
    - **secret value** - the value of the secret, for example, the GitHub token.
    - **description** - the description of the secret in the **Annotations description** field. For example: `GitHub token for the Kyma bot service account`.

2. Apply the `owner` label to the secret in Secret Manager to help identify the secret owner. For example: `owner: neighbors`.
3. Apply the `type` label to the secret in Secret Manager to help identify the secret type. For example: `type: service-account-token`.