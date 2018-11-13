# Kyma Components Migration Guide

## Prerequisites

## Steps
1. Create or adjust Makefile for your component.
Makefile has to define 3 rules:
- **ci-pr** - is executed on PR to master branch
- **ci-master** - is executed on merging PR to master
- **ci-release** - is executed on PR to release branch
Following environment variables will be available when executing Makefile:
- **DOCKER_TAG** - Docker tag
- **DOCKER_PUSH_REPOSITORY** Docker repository where image is be pushed
- **DOCKER_PUSH_DIRECTORY** Repository in Docker repository where image is pushed
2. Create Presubmits and Postsbumit Prow Jobs.
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
>**NOTE:** Yaml file needs to have unique name in the whole `jobs` structure.

## Troubleshooting