# Obligatory security measures

Read about the obligatory security measures to take on a regular basis and when a Kyma organization member leaves the project.

## Change Secrets regularly

All Secrets used in the Prow production cluster must be changed every six months. Follow the [Prow secret management](./prow-secrets-management.md) guidelines to create new Secrets. Once the new Secrets are ready, the [External Secrets Syncer](../../prow/cluster/resources/external-secrets/README.md) changes all Secrets in the Prow cluster automatically.

>**NOTE:** The next Secrets change is planned for October 1, 2020.

## Preventive measures

Make sure that jobs do not include any Secrets that are available in the output as this can lead to severe security issues.

## Offboarding checklist

When a Kyma organization member with access to the Prow cluster leaves the project, take the necessary steps to keep Kyma assets secure.

### Remove Google project access

Remove the person from the `kyma-prow` Google project immediately. Follow [this](https://cloud.google.com/iam/docs/granting-changing-revoking-access) document to revoke necessary access.

### Change Secrets

Change all Secrets that were valid when the person was a project member. Follow the [Prow secret management](./prow-secrets-management.md) guidelines to create new Secrets. Once the new Secrets are ready, the [External Secrets Syncer](../../prow/cluster/resources/external-secrets/README.md) changes all Secrets in the Prow cluster automatically.
