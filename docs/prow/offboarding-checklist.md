# Offboarding Checklist

When someone with access to the Prow cluster leaves the project, we must apply the following steps to keep our assets secured.

## Remove the person from Google project

Remove the person from the `kyma-prow` Google project immediately. Follow [this document](https://cloud.google.com/iam/docs/granting-changing-revoking-access) to revoke access.

## Rotate all the secrets

All the secrets that were valid when the person was in the project should be rotated. Follow [Prow secret management](./prow-secrets-management.md) to create a new key ring and new secrets. Then, use [secrets populator](./../../development/tools/cmd/secretspopulator/README.md) to update all the secrets on Prow cluster.
