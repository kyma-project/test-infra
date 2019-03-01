# Component Jobs

This document describes how to define, modify, and remove standard ProwJobs for `kyma` components. Its purpose is to provide steps required to prepare your component for the Prow CI pipeline.

> **NOTE:** Use the buildpack for Go or Node.js applications provided in the `test-infra` repository. It is the standard mechanism for defining ProwJobs.

When you configure a job for your component, make sure that:

- Your ProwJob is configured to push Docker images to a proper directory. To set it, add the `preset-docker-push-repository` Preset to your ProwJob definition.
- Your ProwJob sends events to GitHub. To configure it, set the **skip_report** parameter to `false`.

## Create component jobs

Creating a CI pipeline is a two-step process:
- You need to define a ProwJob according to steps defined in this section.
- You need to add the job to a release from which you want it to be available. This requires some modifications to the created ProwJob. For details on how to do that, see [this](./release-jobs.md) document which explains how to work with the release jobs.

### Create a presubmit job

Presubmit jobs are jobs that run on pull requests (PRs). They validate changes against the target repository.

Define such a job for your component and store it under the `prow/jobs` directory in the `test-infra` repository.

The structure of this directory looks as follows:

```
jobs
|-- {repository1}
|---- {components}
|------ {component1}
|-------- {component1.yaml}
|------ {component2}
|-------- {component2.yaml}

```

For example, to define a job for the `binding-usage-controller` component from the `kyma` repository, create a `yaml` file called `binding-usage-controller.yaml` in the `test-infra` repository. Place it under the `jobs/kyma/components/binding-usage-controller/` subfolder.

> **NOTE:** All `yaml` files in the whole `jobs` structure need to have unique names.

In your job, call a `build.sh` script from the buildpack image provided in the `test-infra` repository. This script executes the `Makefile` target defined for your component.
The `build.sh` script requires an argument which is the local path to the directory where the `Makefile` for your component is located.

See an example of such a job for the `kyma-project/kyma` repository.

```yaml
presubmits:
  kyma-project/kyma:
    - name: pre-master-kyma-components-binding-usage-controller
      run_if_changed: "^components/binding-usage-controller/"
      branches:
        - master
      skip_report: false
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
              - "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"
            args:
              - "/home/prow/go/src/github.com/kyma-project/kyma/components/binding-usage-controller"
        resources:
          requests:
            memory: 1.5Gi
            cpu: 0.8
```

The table contains the list of all fields in the `yaml` file with their descriptions.

| Field                       | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| --------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **name**                    | The job name. It should clearly identify the job. For reference, see the convention for naming jobs. Add the `pre-master-` prefix in front of presubmit job names and the `post-master-` prefix in front if postsubmit job names for all jobs that run against the `master` branch.                                                                                                                                                                                                                                                                                                                             |
| **run_if_changed**          | A regular expression. Define the component for which changes in a PR must trigger the job. If a PR does not modify the component, this job sends a notification to GitHub with the information that it is skipped.                                                                                                                                                                                                                                                                    |
| **always_run**              | A parameter that defines if a PR must trigger the job automatically. **always_run** and **run_if_changed** are mutually exclusive. If you do not set one of them, you can only trigger the job manually by adding a comment to a PR.                                                                 |
| **branches**                | A list of base branches against which you want the job to run in the PR.                                                                                                                                                                                                                                                                                                                                                                                                              |
| **skip_report**             | A parameter that defines if a job status appears in the PR on GitHub. Set it to `false` for Prow to send the job's status to GitHub.                                                                                                                                                                                                                                                                                        |
| **decorate**                | Decorated jobs automatically clone the repository and store logs from the job execution in Google Cloud Storage (GCS) buckets.                                                                                                                                                                                                                                                                                                                                                        |
| **path_alias**              | The non-standard Go path under which you want to clone the test repository.                                                                                                                                                                                                                                                                                                                                                                                                           |
| **extra_ref**               | Additional repositories to clone in addition to the main repository.                                                                                                                                                                                                                                                                                                                                                                                                                  |
| **spec.containers.image**   | A Docker image. The `test-infra` repository provides Go and Node.js images. For more details, go to `prow/images`. To build a Docker image in your job, you require the privileged mode.                                                                                                                                                                                                                                                                                              |
| **spec.containers.command** | Buildpacks have the `build.sh` script that allows you to easily trigger the build for your component. The `build.sh` script populates some environment variables, such as **DOCKER_TAG**. It also initializes Docker and executes one of the `Makefile` targets, depending on the value of the **BUILD_TYPE** variable. This environment variable can be injected by one of the build's Presets, which are **preset-build-pr**, **preset-build-release**, or **preset-build-master**. |
| **spec.containers.args**    | The `build.sh` script requires a path to the directory where the `Makefile` for your component is located.                                                                                                                                                                                                                                                                                                                                                                        |
| **labels**                  | Regular Kubernetes labels. They are used to enable PodPresets. The available Presets are defined in the `/prow/config.yaml` file. For their detailed descriptions, read the [**Presets**](./presets.md) document.                                                                                                                                                                                                                                                     |
| **spec.resources**          | Regular Kubernetes resources requirements for containers. |

### Check your configuration locally

Use the `development/validate-config.sh` script to validate your Prow configuration. The script accepts three arguments:

- the path to the plugins configuration file (`prow/plugins.yaml`)
- the path to the generic configuration file (`prow/config.yaml`)
- the path to the directory with job definitions (`prow/jobs/`)

See an example:

```bash
cd $GOPATH/src/github.com/kyma-project/test-infra
./development/validate-config.sh prow/plugins.yaml prow/config.yaml prow/jobs/
```

### Review and merge your PR

After your PR is reviewed and approved, merge the changes to the `test-infra` repository. The job configuration is automatically applied on the Prow production cluster. The `config_updater` plugin configured in the `prow/plugins.yaml` file adds a comment to the PR:

![msg](./assets/msg-updated-config.png).

### Create a Makefile for your component

Both buildpacks require a `Makefile` defined in your component directory.  
The `Makefile` has to define these three targets:

- **ci-pr** that is executed for a PR in which the base branch is the `master` branch.
- **ci-master** that is executed after merging a PR to the `master` branch.
- **ci-release** that is executed for a PR issued against the release branch.

See an example of a `Makefile` for the `binding-usage-controller` component:

```Makefile
APP_NAME = binding-usage-controller
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

> **NOTE** Add a tab before each command.

If your job involves pushing a Docker image, its name is based on the following environment variables:

- **DOCKER_TAG** that refers to the Docker tag set by the `build.sh` script.
- **DOCKER_PUSH_DIRECTORY** that points to the directory in the Docker repository where the image is pushed. Set it in the job definition by adding the **preset-build-pr**, **preset-build-master**, or **preset-build-release** Preset.
- **DOCKER_PUSH_REPOSITORY** is the Docker repository where the image is pushed. It is set in the job definition by the **preset-docker-push-repository** Preset.

### Create a PR for your component

The trigger for ProwJobs are GitHub events. For example, GitHub sends events to Prow when you create a PR or add new changes to it.

To test your changes, create a new PR for your component.
If you want to trigger your job again, add a comment on the PR for your component:

- `/test all` to rerun all tests
- `/retest` to only rerun failed tests
- `/test {test-name}` or `/retest {test-name}` to only rerun a specific test. For example, run `/test pre-master-kyma-components-binding-usage-controller`.

After you trigger the job, it appears on `https://status.build.kyma-project.io/`.

### Define a postsubmit job

A postsubmit job is almost the same as the already defined presubmit job for the `master` branch, except for the differences in labels. The postsubmit job uses **preset-build-master** instead of **preset-build-pr**.
To reduce boilerplate code and code repetition, use `yaml` features, such as extending objects.

See an example of the postsubmit job:

```yaml
job_template: &job_template
  skip_report: false
  run_if_changed: "^components/binding-usage-controller/"
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
          - "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"
        args:
          - "/home/prow/go/src/github.com/kyma-project/kyma/components/binding-usage-controller"
        resources:
          requests:
            memory: 1.5Gi
            cpu: 0.8

job_labels_template: &job_labels_template
  preset-dind-enabled: "true"
  preset-sa-gcr-push: "true"
  preset-docker-push-repository: "true"

presubmits: # runs on PRs
  kyma-project/kyma:
    - name: pre-master-kyma-components-binding-usage-controller
      branches:
        - master
      <<: *job_template
      labels:
        <<: *job_labels_template
        preset-build-pr: "true"

postsubmits:
  kyma-project/kyma:
    - name: post-master-kyma-components-binding-usage-controller
      branches:
        - master
      <<: *job_template
      labels:
        <<: *job_labels_template
        preset-build-master: "true"
```

### Define tests for presubmit and postsubmit jobs

To check if your configuration is correct, write a Go test. See the `development/tools/jobs/kyma/binding_usage_controller_test.go` file for reference.
Place your new test under `development/tools/jobs` for the `test-infra-test-jobs-yaml-definitions` presubmit job to execute it.
If you have access to the Prow cluster, there is an option to test a ProwJob on it. For details, see the [official documentation](https://github.com/kubernetes/test-infra/blob/master/prow/build_test_update.md#how-to-test-a-prowjob).


## Modify component jobs

To change the name of the component, follow these steps:

1. Change the **name** value for presubmit and postsubmit jobs for the `master` branch.
2. Change the **run_if_changed** value in **job_template** to a new component path.
3. Create a new release job. For details, see the [official documentation](./release-jobs.md).

>**NOTE**: Do not delete existing release jobs and tests for these jobs as the component you rename is still defined by its old name in previous releases.

4. Modify tests.


## Remove component jobs

To remove a component from Prow, follow these steps:

1. Delete presubmit and postsubmit jobs for the `master` branch.

>**NOTE:** Do not delete jobs for release branches.

2. Delete tests for these jobs.

>**NOTE:** If the component you created is a part of the **X** release, you cannot delete it as the **X.y** release still requires it. For example, a component in `0.6` that is deleted in `0.7` is still needed for `0.6.1`.


## Pipeline overview

To have a better understanding of the role your ProwJob plays in the Prow pipeline, see this flow description:

1. Create a PR that modifies your component.
2. GitHub sends an event to Prow.
3. The `trigger` plugin creates a ProwJob which appears on the `https://status.build.kyma-project.io` page.
4. A Pod is created according to **spec** defined in the presubmit job. Additionally, the decorator clones your repository and mounts it under `/home/prow/go/src/{path_alias}`.
5. The `build.sh` script is executed. It injects the required environment variables and points to the directory passed as an argument. It also executes **make-ci**, **make-master**, or **make-release**, depending on the value of the **BUILD_TYPE** variable.

For further reference, read a more technical insight into the Kubernetes ProwJob flow described in the [**Life of a Prow job**](https://github.com/kubernetes/test-infra/blob/master/prow/life_of_a_prow_job.md) document.
