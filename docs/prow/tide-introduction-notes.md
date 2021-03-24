# Tide introduction

Along with the Prow upgrade we want to introduce Tide for automatically merging the PRs.
Tide is a Prow component that handles merging the PRs once all the requirements are met. **It's a mandatory component for Prow and introducing it is needed in upgrade process.**

### How does tide work?

Every PR introduced to the watched repository will have a new pending context called `tide`. The context will stay in the "pending" state until the PR has passed all the requirements and is in the merge pool.

Tide will sequentially merge the approved PRs one by one.
Tide will do a rebase of a working branch with the latest revision of main branch and then will run the required tests. This ensures the changes are working even after merging.

### Issues

This workflow assumes no-one has write access to the Github repositories. Unfortunately due to some blocking limitations in the `approve` plugin we need to do several workarounds.

- We will still use GitHub's *CODEOWNERS* workflow with manual approvals from the GitHub UI. That means the write access to the repository cannot be revoked. This will require developers **NOT TO** merge PRs manually.
- Tide's context will be set to *Required*, so the PRs can't be merged prematurely. This, however still gives the ability to merge the PR manually once the PR is in the merge pool (tide context is green).
- Skipped jobs will not be reported to the GitHub. `run_if_changed` jobs will be seen as optional to GitHub, but Tide will require them before merging (hence why the tide context is set to required)

### Workflow

1. Developer creates PR
2. After the required jobs have passed the Developer asks for approvals from the code owners.
3. Once the PR has all the required approvals Tide adds PR to the merge pool.
4. Before merge, if the dev branch is older than the latest main tide re-runs the jobs against rebased dev branch.
5. Once the checks have passed again the PR is automatically merged.
6. If the developer does not want the PR to be merged automatically one has to use the `/hold` command which adds the label that blocks automatic merging. To remove this label just simply use `/hold cancel`.

![Tide workflow](./assets/prow-tide-workflow.png)

### Next steps

After upgrading Prow and Tide to the most up-to-date branch we would like to take further steps in increasing development qorkflow quality.

1. Introduce organisational-wide label sync - static set of labels usable by automations and bot command on GitHub across the organisation.
2. Introduce `approve` plugin - with this we'll be able to revoke all the write access from the repositories. Not doable until the `approve` plugin is rewritten.
