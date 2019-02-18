# Prow jobs

This document provides an overview of Prow jobs.  

### Jobs directory structure

Prow jobs reside in the `prow/jobs` directory in the `test-infra` repository. The structure of the `/jobs` directory reflects the Kyma repository structure to make it easier fo you to find jobs you are looking for. 

Job definitions are inluded in `yaml` files, and can be connected to specific components or be more general, like integration jobs. General jobs are defined directly under the `jobs/{repository_name}` directory. Jobs linked to a particular component are available in `jobs/{repository_name}`directory, which includes subdirectories representing each component and containing job definitions. 


> **NOTE:** All `yaml` files in the whole `jobs` structure need to have unique names.


For example:

```
...
prow
|- cluster
| |- starter.yaml
|- images
|- jobs
| |- kyma
| | |- components
| | | |- environments
| | | | |- environments.yaml
| | |- kyma.integration.yaml
|- scripts
|- config.yaml
|- plugins.yaml
...
```

## Job types

You can configure the following job types:

* **Presubmit** jobs run on pull requests (PRs). They validate changes against the target repository. You can 
* **Postsubmit** jobs are almost the same as the already defined presubmit jobs, but they run after the PR is merged. You can notice the difference in labels, as the postsubmit job uses **preset-build-master** instead of **preset-build-pr**.
* **Periodic** jobs run at a specified time, regardless of modifying or merging a PR.

The presubmit and postsubmit jobs for a PR run in random order, and their number depends on the configuration in the `yaml` file. You can check their status on`https://status.build.kyma-project.io/`.


### Naming convention 

When you define jobs for Prow, the **name** parameter of the job must follow one of these patterns:

  - `{prefix}-{repository-name}-{component-name}-{job-name}` for components
  - `{prefix}-{repository-name}-{job-name}` for jobs not connected to a particular component

Add `{prefix}` in front of all presubmit and postsubmit jobs. Use:
- `pre-master` for presubmit jobs that run against the `master` branch.
- `post-master` for postsubmit jobs that run against the `master` branch.
- `pre-rel{release-number}` for presubmit jobs that run against the release branches. For example, write `pre-rel06-kyma-components-api-controller`.

In both cases, `{job_name}` must reflect the job's responsibility.


## Triggers

Prow runs presubmit and postsubmit jobs based on the following parameters: 

* `always_run: true` means that the job run automatically at all times.
* `run_if_changed: true` means that the job runs automatically each time you change files located in a specific directory. 

If you set the **always_run** parameter to `false` to and leave `run_if_changed` without any value, the job won't run unless you trigger it manually.


## Interact with Prow

You can interact with Prow to trigger the failed jobs again. 

> **NOTE:** You can rerun only presubmit jobs.

If you want to trigger your job again, add a comment on the PR for your component:

`/test all` to rerun all tests
`/retest` to only rerun failed tests
`/test {test-name}` or `/retest {test-name}` to only rerun a specific test. For example, run `/test pre-master-kyma-components-binding-usage-controller`.


## Create jobs

For details on how to create jobs see:
* [Create component jobs](./component-jobs.md)
* [Create release jobs](./release-jobs.md)
