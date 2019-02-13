# Authorization

## Required GCP Permissions

## Kubernetes RBAC rules on Prow cluster

## User permissions on GitHub

Prow is responsible for starting tests in reaction to certain Github events. For security reasons, the `trigger` plugin ensures that test jobs are run only on pull requests created or verified by trusted users.

### Trusted users
Members of the `kyma-project` organization are considered trusted users. Trigger starts jobs automatically when a trusted user opens a pull request or commits changes to a pull request branch. Alternatively, trusted collaborators can start jobs manually via the `/test all`, `/test {JOB_NAME}` and `/retest` commands, even if a particular pull request was created by an external user. 

### External contributors
External contributors are users outside the `kyma-project` organization. Trigger does not automatically start test jobs on pull requests created by external contributors. Furthermore, external contributors are not allowed to manually run tests on their own pull requests.

> **NOTE:** External contributors can still trigger tests on pull requests created by trusted users.


## Authorization decisions enforced by Prow