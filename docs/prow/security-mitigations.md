# Security Mitigations

## Regularly Rotate all the Secrets

Next rotation date: **01.08.2019**

All the secret used in the production Prow cluster must be rotated regularly. Rotation interval is six months. Follow the [Prow secret management](./prow-secrets-management.md) to create a new keyring and new Secrets. Then, use [Secrets populator](./../../development/tools/cmd/secretspopulator/README.md) to update all the Secrets in the Prow cluster.

## Do not Print Out Secrets

Developers must not write jobs that emit secrets to the output. This can lead to severe problems.

## Offboarding Checklist

When someone with access to the Prow cluster leaves the project, we must apply the following steps to keep our assets secured.

### Remove the person from the Google project

Remove the person from the `kyma-prow` Google project immediately. Follow [this document](https://cloud.google.com/iam/docs/granting-changing-revoking-access) to revoke access.

### Rotate all the Secrets

All the Secrets that were valid when the person was in the project must be rotated. Follow the [Prow secret management](./prow-secrets-management.md) to create a new keyring and new Secrets. Then, use [Secrets populator](./../../development/tools/cmd/secretspopulator/README.md) to update all the Secrets in the Prow cluster.
