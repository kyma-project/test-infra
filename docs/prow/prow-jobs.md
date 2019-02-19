# Prow jobs

This document provides an overview of ProwJobs.  

## Directory structure

ProwJobs reside in the `prow/jobs` directory in the `test-infra` repository. Job definitions are configured in `yaml` files, and can be connected to specific components or be more general, like for integration jobs. General jobs are defined directly under the `jobs/{repository_name}` directory. Jobs configured for components are available in `jobs/{repository_name}`directory which includes subdirectories representing each component and containing job definitions. 


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

- **Presubmit** jobs run on pull requests (PRs). They validate changes against the target repository. By default, all presubmit jobs are required to pass before you can merge the PR.  If you set the **optional** parameter to `true`, a job becomes optional and you can still merge your PR even if the job fails. 
- **Postsubmit** jobs are almost the same as the already defined presubmit jobs, but they run after the PR is merged. You can notice the difference in labels, as the postsubmit job uses **preset-build-master** instead of **preset-build-pr**.
- **Periodic** jobs run automatically at a scheduled time. You don't need to modify or merge the PR to trigger them. 

The presubmit and postsubmit jobs for a PR run in random order, and their number for a PR depends on the configuration in the `yaml` file. You can check the job status on [`https://status.build.kyma-project.io/`](https://status.build.kyma-project.io/).


## Naming convention 

When you define jobs for Prow, the **name** parameter of the job must follow one of these patterns:

- `{prefix}-{repository-name}-{path-to-component}-{job-name}` for components
- `{prefix}-{repository-name}-{job-name}` for jobs not connected to a particular component


Add `{prefix}` in front of all presubmit and postsubmit jobs. Use:
- `pre-master` for presubmit jobs that run against the `master` branch.
- `post-master` for postsubmit jobs that run against the `master` branch.
- `pre-rel{release-number}` for presubmit jobs that run against the release branches. For example, write `pre-rel06-kyma-components-api-controller`.

In both cases, `{job_name}` must reflect the job's responsibility.


## Triggers

Prow runs presubmit and postsubmit jobs based on the following parameters: 

- `always_run: true` for the job to run automatically at all times.
- `run_if_changed: {regular expression}` for the job to run if a PR modifies files matching the pattern. If a PR does not modify the files, this job sends a notification to GitHub with the information that it is skipped.

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

- [Create component jobs](./component-jobs.md)
- [Create release jobs](./release-jobs.md)

For further details of ProwJobs see [this](https://github.com/kubernetes/test-infra/blob/master/prow/jobs.md) document.