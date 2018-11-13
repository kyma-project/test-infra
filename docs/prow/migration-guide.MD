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
Presubmit job runs on a Pull Request (PR). Postsubmit job runs on merging to a branch.
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

For example, if you define jobs for `helm-broker` component from `kyma` repository,
you create one file called: `jobs/kyma/components/helm-broker/helm-broker.jobs.yaml`


|field                  | Value | Description|
|-----------------------|-------------|---------------|
|optional               | TODO        | TODO
|skip_report            | true        | If true, job's status will not be send to GitHub. 
|decorate               | true        | Decorated job provide following functionalities: automatic repository cloning, storing logs in a buckets
|path_alias             | -           | Where the repository under test is cloned
|extra_ref              | -           | Auxiliary repositories that need to be cloned
|spec.containers.image  |             | Docker image. Currently provided images are buildpack-golang and buildpack-node. For more details see `prow/images`. Both of them required privileged mode.  
|spec.containers.command|             | Buildpacks have a `pipeline.sh` script that allows easily trigger build for your component. Pipeline script populate all required environment variables, and then execute one of the makefile rules (depending on the environment variable **BUILD_TYPE**)
|spec.containers.env    |             | **SOURCE_DIR** is requred by `pipeline.sh` script and points to the directory where Makefile for your component is placed
|name                   |             | Job name. See TODO for requirements for job name
|context                |             | GitHub status context for the job. Should be the same as job name
|branches               | master      | Run it only for PR's for that branches
|run_if_changed         |             | Regular expression, run this build only if this files were changed. In case of other files changed, this job will send a notification to GitHub that it is skipped.
|labels                 |             | Regular k8s labels. In this case we use it to enable podpresets. Available presets can be fonund in `/prow/config.yaml` file.

## Available presets
| name                       | description
|----------------------------|------------
| preset-gc-compute-envs        | Provide environment variables with gcloud compute zone and gclod compute region
| preset-gc-project-env         | Provide environment variable with gcloud project name
| preset-sa-vm-kyma-integration | Injects credentials for Service Account for running integration tests on VM
| preset-sa-gke-kyma-integration| Injects credentials for Service Account for running integration tests on GKE cluster
| preset-sa-gcr-push            | Injects credentials for pushing images to Google Cloud Registry
| preset-dind-enabled           | Allows to run docker in your docker container
| preset-docker-push-repository | Provide environment variable with address of Docker repository
| preset-build-pr               | Provide environment variable with directory in Docker repository to store images and variable **BUILD_TYPE** set to **pr** 
| preset-build-master           | As above, but **BUILD_TYPE** is set to **master**
| preset-build-release          | As above, but **BUILD_TYPE** is set to **release**.



See example:
```yaml
job_template: &job_template
  optional: true 
  skip_report: true # if true, job will be triggered, by it's status will not be send to GitHub
  decorate: true # decorated jobs provide following functionalities: cloning repositories, storing logs in a buckets
  path_alias: github.com/kyma-project/kyma
  extra_refs: # if we need another repository to be cloned, we can provide it here
  - org: kyma-project
    repo: test-infra
    base_ref: master
    path_alias: github.com/kyma-project/test-infra
  spec:
    containers:
    - image: eu.gcr.io/kyma-project/prow/buildpack-golang:0.0.1 # use Golang buildpack
      securityContext:
        privileged: true
      command:
      # pipeline script populate all required environment variables, and then execute 
      # one of the makefile rules (depending on the environment variable BUILD_TYPE
      - "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/pipeline.sh" 
      env:
      # SOURCE_DIR points to the directory where is a Makefile for your component
      - name: SOURCES_DIR
        value: "/home/prow/go/src/github.com/kyma-project/kyma/components/ui-api-layer"



job_labels_template: &job_labels_template
  preset-dind-enabled: "true"
  preset-sa-gcr-push: "true"
  preset-docker-push-repository: "true"

presubmits: # runs on PRs
  kyma-project/kyma:
  - name: &name prow/kyma/components/ui-api-layer
    context: *name #name and context is the same
    branches:
    - master # run it only for PR for master branch
    # regular expression, run this build only if this files were changed. 
    # In case of other files changed, this job will send a notifaction to GitHub that it is skipped.
    run_if_changed: "components/ui-api-layer/"  
    labels:
      # labels are used for enabling presets. 
      # Available presets can be fonund in /prow/config.yaml file.
      preset-build-pr: "true"
      <<: *job_labels_template
    <<: *job_template
  - name: &name prow/release/kyma/components/ui-api-layer
    context: *name
    run_if_changed: "(components/ui-api-layer/|resources/core/values.yaml)"
    branches:
    - '^release-\d+\.\d+$'
    labels:
      preset-build-release: "true"
      <<: *job_labels_template
    <<: *job_template
```

>**NOTE:** Yaml file needs to have unique name in the whole `jobs` structure.

### Review and merge your PR. 
After merging, job configuration is automatically applied on a Prow production cluster 
and comment is added to the PR:
![msg](./assets/msg-updated-config.png)


## Flow