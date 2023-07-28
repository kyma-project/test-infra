# Obligatory security measures

Read about the obligatory security measures to take on a regular basis and when a Kyma organization member leaves the project.

## Preventive measures

Make sure that jobs do not include any Secrets that are available in the output as this can lead to severe security issues.

## Offboarding checklist

When a Kyma organization member with access to the Prow cluster leaves the project, take the necessary steps to keep Kyma assets secure.

### Remove Google project access

Remove the person from the `kyma-prow` Google project immediately. Follow [this](https://cloud.google.com/iam/docs/granting-changing-revoking-access) document to revoke necessary access.
