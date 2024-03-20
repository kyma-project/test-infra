# How to name a secret

This tutorial describes how to name a secret in the Google Secret Manager.

## Naming convention

1. The secret name should be in the following format: `<tool>_<component>_<entity>`. For example: `security-backend_publisher_github-kyma-bot-token`. Where `security-backend` is the tool, component is `publisher` and `github-kyma-bot-token` is the entity. The component part is optional and can be skipped if the tool does not contain multiple components.
   From the name it should be clear what the secret contains and is used, a GitHub token for the Kyma bot service account which is used in the Publisher component of the  Security Backend tool.
   The same secret should not have two entries in the Secret manager with different names. For example, the `prow_default_sap-slack-bot-token` and `workloads_default_sap-slack-bot-token` should. It should be only one entry in the Secret Manager with the name `prow_notifier_slack-bot-token`.
   The secret should contain the following fields:
    - **secret value** - the value of the secret, for example, the GitHub token.
    - **description** - the description of the secret in the **Annotations description** field. For example: `GitHub token for the Kyma bot service account which is used in the Security Dashboard tool`.

2. Apply the `owner` label to the secret in Secret Manager to help identify the secret owner. For example: `owner: neighbors`.
3. Apply the `type` label to the secret in Secret Manager to help identify the secret type. For example: `type: service-account-token`.
4. Apply the `tool` label to the secret in Secret Manager to help identify the tool where secret is used. For example: `tool: security-backend`.
5. Apply the `component` label to the secret in Secret Manager to help identify the component of the tool where secret is used. For example: `component: publisher`.
6. Apply the `entity` label to the secret in Secret Manager to help identify the entity of secret. For example: `entity: github-kyma-bot-token`.