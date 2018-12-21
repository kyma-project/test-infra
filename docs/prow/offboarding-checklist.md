# Offboarding Checklist

When someone with access to the Prow cluster leaves the project, we must apply the following steps to keep our assets secured.

## Remove the person from the Google project

Remove the person from the `kyma-prow` Google project immediately. Follow [this document](https://cloud.google.com/iam/docs/granting-changing-revoking-access) to revoke access.

## Rotate all the Secrets

All the Secrets that were valid when the person was in the project must be rotated. Follow the [Prow secret management](./prow-secrets-management.md) to create a new keyring and new Secrets. Then, use [Secrets populator](./../../development/tools/cmd/secretspopulator/README.md) to update all the Secrets in the Prow cluster.
