# Kyma Components Migration Guide

This document describes procedure for defining Prow jobs for `kyma` component. In this document we assume that 
you are using standard mechanism provided by `test-infra` respository for defining prowjobs, that is to say you are using
provided buildpack for Go or node application.


## Steps
### Create or adjust the `Makefile` for your component.
Both of the buildpacks require a `Makefile` defined for your component in `kyma-project/kyma` repository.   
The `Makefile` has to define 3 rules:
- **ci-pr** - is executed for a Pull Request(PR) to the `master` branch.
- **ci-master** - is executed after merging a PR to the `master` branch.
- **ci-release** - is executed for a PR to the release branch.

Following environment variables are available when executing Makefile:
- **DOCKER_TAG** - Docker tag.
- **DOCKER_PUSH_REPOSITORY** - Docker repository where image is be pushed.
- **DOCKER_PUSH_DIRECTORY** directory in Docker repository where image is pushed.

Example Makefile for `ui-api-layer` component:
```Makefile
IMG_NAME = ui-api-layer
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(IMG_NAME)
TAG = $(DOCKER_TAG)

ci-pr: resolve build-and-test build-image push-image
ci-master: resolve build-and-test build-image push-image
ci-release: resolve build-and-test build-image push-image

resolve:
	dep ensure -v -vendor-only
build-and-test:
	./before-commit.sh ci
build-image:
	docker build -t $(IMG_NAME) .
push-image:
	docker tag $(IMG_NAME) $(IMG):$(TAG)
	docker push $(IMG):$(TAG) 
```

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

#### Create presubmit job for master branch
```yaml
presubmits: # runs on PRs
  kyma-project/kyma:
    - name: prow/kyma/components/ui-api-layer
      run_if_changed: "components/ui-api-layer/"
      context: prow/kyma/components/ui-api-layer
      branch: master
      skip_report: true # from documentation: SkipReport skips commenting and setting status on GitHub.
      decorate: true
      path_alias: github.com/kyma-project/kyma
      extra_refs:
        - org: kyma-project
          repo: test-infra
          base_ref: master
          path_alias: github.com/kyma-project/test-infra
      labels:
        preset-dind-enabled: "true"
        preset-sa-gcr-push: "true"
        preset-docker-push-repository: "true"
        preset-build-pr: "true"
      spec:
        containers:
          - image: eu.gcr.io/kyma-project/prow/buildpack-golang:0.0.1
            securityContext:
              privileged: true
            command:
              - "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/pipeline.sh"
            env:
              - name: SOURCES_DIR
                value: "/home/prow/go/src/github.com/kyma-project/kyma/components/ui-api-layer"

```
Fields description:

|Field                  | Value | Description|
|-----------------------|-------------|---------------| 
|name                   |             | Job name. Should unambigously identifies job. See [TODO]() requirements for job name.
|run_if_changed         |             | Regular expression. Define to run this build only if this files were changed. In case of other files changed, this job will send a notification to GitHub that it is skipped.
|context                |             | GitHub status context for the job. Should be the same as job name.
|branches               | master      | Run it only for PR's for that branches
|skip_report            | true        | If true, job's status will not be send to GitHub. 
|decorate               | true        | Decorated jobs clones automatically repository, and store logs from the execution in a buckets
|path_alias             | -           | Where the repository under test is cloned.
|extra_ref              | -           | Auxiliary repositories that need to be also cloned, in addition main repository. 
|spec.containers.image  |             | Docker image. Currently provided images are buildpack-golang and buildpack-node. For more details see `prow/images`. Both of them required privileged mode.  
|spec.containers.command|             | Buildpacks have a `pipeline.sh` script that allows easily trigger build for your component. Pipeline script populate all required environment variables, and then execute one of the makefile rules (depending on the environment variable **BUILD_TYPE**)
|spec.containers.env    |             | **SOURCE_DIR** is requred by `pipeline.sh` script and points to the directory where Makefile for your component is placed
|labels                 |             | Regular k8s labels. In this case we use it to enable podpresets. Available presets can be fonund in `/prow/config.yaml` file and are described in section Available Presets.

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





>**NOTE:** Yaml file needs to have unique name in the whole `jobs` structure.

### Review and merge your PR. 
After merging, job configuration is automatically applied on a Prow production cluster 
and comment is added to the PR:
![msg](./assets/msg-updated-config.png)

## Testing
There are following options to tirgger your job:
- create a new PR
- request executing all tests for already created PR by adding comment: `/test all`

## Define postsubmit job and presubmit job for release
Postsubmit job and presubmit job for releases are almost the same as already defined job for master,
except for the differences in:
- Branch name.
- **run_if_changed** parameters.
- Labels, since you use `preset-build-release` instead of `preset-build-pr`.
To reduce boilerplate code and code repetition, use `yaml` features, such as extending objects.
See example:
```yaml
job_template: &job_template
  optional: true 
  skip_report: true
  decorate: true 
  path_alias: github.com/kyma-project/kyma
  extra_refs: 
  - org: kyma-project
    repo: test-infra
    base_ref: master
    path_alias: github.com/kyma-project/test-infra
  spec:
    containers:
    - image: eu.gcr.io/kyma-project/prow/buildpack-golang:0.0.1 
      securityContext:
        privileged: true
      command:

      - "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/pipeline.sh" 
      env:
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
    - master
    run_if_changed: "components/ui-api-layer/"  
    labels:
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

To check if your configuration is proper, you can write a Go test, for a reference see `prow/jobs/kyma/components/ui-api-layer/jobs_test.go`.
## Flow