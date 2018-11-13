# Kyma Components Migration Guide

## Prerequisites

## Steps
### Create or adjust the `Makefile` for your component.
The `Makefile` has to define 3 rules:
- **ci-pr** - is executed for a Pull Request(PR) to the `master` branch.
- **ci-master** - is executed after merging a PR to the `master` branch.
- **ci-release** - is executed for a PR to the release branch.

Following environment variables are available when executing Makefile:
- **DOCKER_TAG** - Docker tag.
- **DOCKER_PUSH_REPOSITORY** - Docker repository where image is be pushed.
- **DOCKER_PUSH_DIRECTORY** directory in Docker repository where image is pushed.

### Create presubmits and postsbumit Prow jobs.
All jobs needs to be stored in `prow/jobs` directory. 
Structure of this directory is following:
```
jobs
|-- <repository1>
|---- components
|------ <component1>
|-------- <component1.yaml>
|------ <component2>

```

For example, if you define job for `helm-broker` component from `kyma` repository,
you create file `jobs/kyma/components/helm-broker/helm-broker.jobs.yaml`

TODO: how to write a job

>**NOTE:** Yaml file needs to have unique name in the whole `jobs` structure.

### Review and merge your PR. 
After merging, job configuration is automatically applied on a Prow production cluster 
and comment is added to the PR:
![msg](./assets/msg-updated-config.png)


## Troubleshooting