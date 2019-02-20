# ProwJobs

This document provides an overview of ProwJobs.  

## Directory structure

ProwJobs reside in the `prow/jobs` directory in the `test-infra` repository. Job definitions are configured in `yaml` files. ProwJobs can be connected to specific components or be more general, like for integration jobs. General jobs are defined directly under the `jobs/{repository_name}` directories. Jobs configured for components are available in `jobs/{repository_name}`directories which include subdirectories representing each component and containing job definitions. 


This is a sample directory structure:

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
> **NOTE:** All `yaml` files in the whole `jobs` structure need to have unique names.

## Job types

You can configure the following job types:

- **Presubmit** jobs run on pull requests (PRs). They validate changes against the target repository. By default, all presubmit jobs must pass before you can merge the PR.  If you set the **optional** parameter to `true`, a job becomes optional and you can still merge your PR even if the job fails. 
- **Postsubmit** jobs are almost the same as the already defined presubmit jobs, but they run when you merge the PR. You can notice the difference in labels as postsubmit jobs use **preset-build-master** instead of **preset-build-pr**.
- **Periodic** jobs run automatically at a scheduled time. You don't need to modify or merge the PR to trigger them. 

The presubmit and postsubmit jobs for a PR run in a random order. Their number in a PR depends on the configuration in the `yaml` file. You can check the job status on [`https://status.build.kyma-project.io/`](https://status.build.kyma-project.io/).


## Naming convention 

When you define jobs for Prow, the **name** parameter of the job must follow one of these patterns:

- `{prefix}-{repository-name}-{path-to-component}` for components
- `{prefix}-{repository-name}` for jobs not connected to a particular component

You can extend the name of the job with a suffix to indicate the job's purpose. For example, write `pre-master-kyma-integration`.

Add `{prefix}` in front of all presubmit and postsubmit jobs. Use:
- `pre-master` for presubmit jobs that run against the `master` branch.
- `post-master` for postsubmit jobs that run against the `master` branch.
- `pre-rel{release-number}` for presubmit jobs that run against the release branches. For example, write `pre-rel06-kyma-components-api-controller`.




## Triggers

Prow runs presubmit and postsubmit jobs based on the following parameters: 

- `always_run: true` for the job to run automatically at all times.
- `run_if_changed: {regular expression}` for the job to run if a PR modifies files matching the pattern. If a PR does not modify the files, this job sends a notification to GitHub with the information that the job is skipped.

**always_run** and **run_if_changed** are mutually exclusive. If you do not set one of them, you can only trigger the job manually by adding a comment to a PR.                                                               


## Interact with Prow

Prow allows you to use comments to rerun presubmit jobs on PRs.

> **NOTE:** You can rerun only presubmit jobs.

If you want to trigger your job again, add one of these comments to your PR:

`/test all` to rerun all tests
`/retest` to only rerun failed tests
`/test {test-name}` or `/retest {test-name}` to only rerun a specific test. For example, run `/test pre-master-kyma-components-binding-usage-controller`.


## Create jobs

For details on how to create jobs, see:

- [Create component jobs](./component-jobs.md)
- [Create release jobs](./release-jobs.md)

For further reference, read a more technical insight into the Kubernetes [ProwJobs](https://github.com/kubernetes/test-infra/blob/master/prow/jobs.md).