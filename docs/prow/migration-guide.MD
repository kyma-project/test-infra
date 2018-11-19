# Define standard jobs for Kyma components

This document describes the procedure for defining Prow jobs for `kyma` component. In this document, we assume that 
you are using the standard mechanism provided by `test-infra` repository for defining prowjobs, that is to say, you are using
provided buildpack for Go or node application.

## Migration plan
In the first phase, we aim to have duplicated jobs for components - in the internal CI and in the Prow. 
This is a temporary state until we ensure that our Prow server is production-ready. Hence:
- Prow jobs should be configured to push Docker images to a different directory than internal CI is pushing. 
This is achieved by a preset `preset-docker-push-repository`.
- Prow jobs should not send events to GitHub (`skip_report` set to `true`).

## Steps

### Create presubmit job
The presubmit job runs on a Pull Request (PR). All jobs need to be stored in `test-infra` repository, in `prow/jobs` directory.
Structure of this directory is following:
```
jobs
|-- <repository1>
|---- components
|------ <component1>
|-------- <component1.yaml>
|------ <component2>

```

For example, if you define jobs for `binding-usage-controller` component from `kyma` repository,
you create one file called: `jobs/kyma/components/binding-usage-controller/binding-usage-controller.yaml`
>**NOTE:** YAML file needs to have a unique name in the whole `jobs` structure.

In your job, you should call a `pipeline.sh` script from provided buildpack image. This script executes makefile's target defined for your component.
Script `pipeline.sh` requires the environment variable `SORUCES_DIR` which is a path to a directory where a `Makefile` for your component is defined. 

The content of the job file should be the following:
```yaml
presubmits: # runs on PRs
  kyma-project/kyma:
    - name: prow/kyma/components/binding-usage-controller
      run_if_changed: "^components/binding-usage-controller/"
      context: prow/kyma/components/binding-usage-controller
      branch: master
      skip_report: true
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
          - image: eu.gcr.io/kyma-project/prow/test-infra/buildpack-golang:v20181119-afd3fbd
            securityContext:
              privileged: true
            command:
              - "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/pipeline.sh"
            env:
              - name: SOURCES_DIR
                value: "/home/prow/go/src/github.com/kyma-project/kyma/components/binding-usage-controller"

```

As you can see, we defined one job, for `kyma-project/kyma` repository.
 
Fields description:

|Field                  | Description|
|-----------------------|-------------|
|name                   | Job name. Should unambiguously identify the job. See [convention](./../../prow/README.md#convention-for-naming-jobs) for job name.
|run_if_changed         | Regular expression. Define to run this build only if these files were changed. In case of PR that does not modify your component, this job sends a notification to GitHub that it is skipped.
|context                | GitHub status context for the job - how it be displayed on the GitHub. Should be the same as the job name.
|branches               | Run it only for PR's that base branch appears on this list
|skip_report            | If true, job's status will not be sent to GitHub. Set it to `true`.
|decorate               | Decorated jobs clones automatically repository and store logs from its execution in Google Cloud Storage buckets.
|path_alias             | Where the repository under test is cloned.
|extra_ref              | Auxiliary repositories that need to be also cloned, in addition to the main repository. 
|extra_ref              | Auxiliary repositories that need to be also cloned, in addition to the main repository. 
|spec.containers.image  | Docker image. Currently provided images are buildpack-golang and buildpack-node. For more details see `prow/images`. If you build docker image in your job, the privileged mode is required. 
|spec.containers.command| Buildpacks have a `pipeline.sh` script that allows easily trigger build for your component. `pipeline.sh` script populates some environment variables(**DOCKER_TAG), initializes Docker, and then execute one of the makefile targets depending on the environment variable **BUILD_TYPE**. This environment variable can be injected by one of the build's presets (preset-build-pr,preset-build-release or preset-build-master).
|spec.containers.env    | **SOURCE_DIR** is required by the `pipeline.sh` script and points to the directory where Makefile for your component is placed.
|labels                 | Regular k8s labels. In this case, we use it to enable podpresets. Available presets can be found in `/prow/config.yaml` file and are described below.

#### Available presets
| name                       | description
|----------------------------|------------
| preset-dind-enabled           | Allows running Docker in your job
| preset-sa-gcr-push            | Injects credentials for pushing images to Google Cloud Registry
| preset-docker-push-repository | Provide environment variable with the address of Docker repository
| preset-build-pr               | Provide environment variable with the directory in Docker repository to store images and variable **BUILD_TYPE** set to **pr** 
| preset-build-master           | As above, but **BUILD_TYPE** is set to **master**
| preset-build-release          | As above, but **BUILD_TYPE** is set to **release**.
| preset-gc-project-env         | Provide environment variable with Gcloud project name
| preset-gc-compute-envs        | Provide environment variables with the Gcloud compute zone and the Gcloud compute region
| preset-sa-vm-kyma-integration | Injects credentials for Service Account for running integration tests on VM
| preset-sa-gke-kyma-integration| Injects credentials for Service Account for running integration tests on GKE cluster



### Check your configuration locally
Use `development/validate-config.sh` script to validate Prow configuration. The script accepts 3 arguments:
- path to plugins configuration file
- path to generic configuration file
- path to directory with job definitions

```bash
cd $GOPATH/src/github.com/kyma-project/test-infra
./development/validate-config.sh prow/plugins.yaml prow/config.yaml prow/jobs/
Checking plugin configuration from 'prow/plugins.yaml' and prow configuration from 'prow/config.yaml and jobs configuration from 'prow/jobs/'
```
 
### Review and merge your PR on test-infra
After merging, job configuration is automatically applied on a Prow production cluster 
and comment is added to the PR:
![msg](./assets/msg-updated-config.png).
This is done by `config_updater` configured in `prow/plugins.yaml` file.

 
### Create or adjust the `Makefile` for your component.
Both of the buildpacks require a `Makefile` defined in your component directory.   
The `Makefile` has to define 3 targets:
- **ci-pr** - is executed for a Pull Request(PR) that base branch is the `master`.
- **ci-master** - is executed after merging a PR to the `master` branch.
- **ci-release** - is executed for a PR to the release branch.

Example Makefile for `binding-usage-controller` component:
```Makefile
APP_NAME = "binding-usage-controller"
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG = $(DOCKER_TAG)
binary=$(APP_NAME)

.PHONY: build
build:
	./before-commit.sh ci

.PHONY: build-image
build-image:
	cp $(binary) deploy/controller/$(binary)
	docker build -t $(APP_NAME):latest deploy/controller

.PHONY: push-image
push-image:
	docker tag $(APP_NAME) $(IMG):$(TAG)
	docker push $(IMG):$(TAG)

.PHONY: ci-pr
ci-pr: build build-image push-image

.PHONY: ci-master
ci-master: build build-image push-image

.PHONY: ci-release
ci-release: build build-image push-image

.PHONY: clean
clean:
	rm -f $(binary)

```
>**NOTE** Tab before each command is required.

It is highly likely, that as a result of your job, new Docker image will be pushed. 
In the above example, you can see that image name is deduced on the  basis of the following environment variables:
- **DOCKER_TAG** - Docker tag. (set by `pipeline.sh` script)
- **DOCKER_PUSH_DIRECTORY** directory in Docker repository where a image is pushed. (set in job definition by one of the preset `preset-build-pr`, `preset-build-master`,`preset-build-release`)
- **DOCKER_PUSH_REPOSITORY** - Docker repository where image is be pushed. (set in job definition by the preset: `preset-docker-push-repository`)

### Create a PR for your component
ProwJobs usually are triggered by GitHub events. Events are sent, for example when a PR is created, or 
new changes were added to a PR. To test your changes, create a new PR for your component. 
If you want to trigger your job again, add a comment on PR for your component:
- `/test all` - to run all tests
- `/restest` - to run again only failed tests 
- `/test {your test name}`, for example `/test prow/kyma/components/binding-usage-controller` to run only a specific test
Then your job should appear on `https://status.build.kyma-project.io/`.

### Define postsubmit job 
The postsubmit job is almost the same as the already defined presubmit job for the `master`, except for the differences in:
- Labels, since you use `preset-build-master` instead of `preset-build-pr`.
To reduce boilerplate code and code repetition, use YAML features, such as extending objects.
See example:
```yaml
job_template: &job_template
  optional: true
  skip_report: true
  decorate: true
  path_alias: github.com/kyma-project/kyma
  max_concurrency: 10
  extra_refs:
  - org: kyma-project
    repo: test-infra
    base_ref: master
    path_alias: github.com/kyma-project/test-infra
  spec:
    containers:
    - image: eu.gcr.io/kyma-project/prow/test-infra/buildpack-golang:v20181119-afd3fbd
      securityContext:
        privileged: true
      command:
      - "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/pipeline.sh"
      env:
      - name: SOURCES_DIR
        value: "/home/prow/go/src/github.com/kyma-project/kyma/components/binding-usage-controller"

job_labels_template: &job_labels_template
  preset-dind-enabled: "true"
  preset-sa-gcr-push: "true"
  preset-docker-push-repository: "true"

presubmits: # runs on PRs
  kyma-project/kyma:
  - name: &name prow/kyma/components/binding-usage-controller
    run_if_changed: "^components/binding-usage-controller/"
    context: *name
    branches:
      - master
    <<: *job_template
    labels:
      <<: *job_labels_template
      preset-build-pr: "true"

postsubmits:
  kyma-project/kyma:
    - name: &name prow/kyma/components/binding-usage-controller
      context: *name
      branches:
        - master
      <<: *job_template
      labels:
        <<: *job_labels_template
        preset-build-master: "true"

```

To check if your configuration is proper, you can write a Go test, for a reference see `development/tools/jobs/buc_test.go`.

## Full story 
1. Create PR that modifies your component
2. GitHub sends an event to Prow
3. ProwJob is created and appears on `http://status.build.kyma-project.io` page.
4. Pod according to `Spec` is created. Additionally, decorator clones your repository and mount it to `/home/prow/go/src/{path_alias}`
5. The `pipeline.sh` script is executed. It injects required environment variables, then go to directory defined by the **SOURCES_DIR** variable, and execute `make-ci`, `make-master` or `make-release` depending on the value of **BUILD_TYPE** environment variable.

If you need more technical explanation, familiarize with [this document](https://github.com/kubernetes/test-infra/blob/master/prow/life_of_a_prow_job.md).