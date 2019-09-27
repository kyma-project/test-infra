# Release Jobs

This document describes the procedure for defining release jobs for Kyma components.

>**NOTE:** Before you follow the steps in this guide, read the [**Component jobs**](./component-jobs.md) document to learn how to create a standard component job. To learn about the release process, read the [**Release process**](https://github.com/kyma-project/community/blob/master/guidelines/releases/release-process.md) document that guides you through the steps required to prepare and create a Kyma release.

## Steps

Follow the subsections to define component jobs for a release.

### Implement a release rule
Define the `ci-release` released rule in your component's Makefile.
See the Binding Usage Controller as an example:
```
APP_NAME = binding-usage-controller
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG = $(DOCKER_TAG)
binary=$(APP_NAME)

...

.PHONY: ci-release
ci-release: build build-image push-image

```

### Define a release job

The differences between a release job and a job for the `master` branch are as follows:
- Different branches
- Different prefixes for job names. Use `pre-rel{number}-` in front of the given release job name. For example, write `pre-rel06-kyma-components-binding-usage-controller`.
- The `preset-build-release` label used instead of `preset-build-pr`
- The **extra_refs** parameter for the `test-infra` repository that uses the `release-0.6` branch instead of `master`
- The **always_run** parameter set to `true` instead of specifying the **run_if_changed** parameter

See an example:
```yaml
test_infra_ref: &test_infra_ref
  org: kyma-project
  repo: test-infra
  path_alias: github.com/kyma-project/test-infra

job_template: &job_template
  name: pre-rel06-kyma-components-binding-usage-controller
  skip_report: false
  decorate: true
  path_alias: github.com/kyma-project/kyma
  max_concurrency: 10
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
  - branches:
    - ^master$
    <<: *job_template
    run_if_changed: "^components/binding-usage-controller/"
    extra_refs:
    - <<: *test_infra_ref
      base_ref: master
    labels:
      <<: *job_labels_template
      preset-build-pr: "true"
  - branches:
    - release-0.6
    <<: *job_template
    always_run: true
    extra_refs:
    - <<: *test_infra_ref
      base_ref: release-0.6
    labels:
      <<: *job_labels_template
      preset-build-release: "true"

postsubmits:
  kyma-project/kyma:
  - branches:
    - ^master$
    <<: *job_template
    run_if_changed: "^components/binding-usage-controller/"
    extra_refs:
    - <<: *test_infra_ref
      base_ref: master
    labels:
      <<: *job_labels_template
      preset-build-master: "true"

```

The component job configuration in this guide differs from the one defined in the [**Component jobs**](./component-jobs.md) document as follows:

- The **test-infra-ref** object is defined, where **org**, **repo**, and **path_alias** are specified.
- **job-template** now defines **name**, but **run_if_changed** and **extra_refs** are removed from it.
**run_if_changed** is defined only for the presubmit and postsubmit job for the `master` branch.
- All jobs must define proper **extra_refs** with the specified **base_ref**.
- Every job uses a different build Preset (**preset-build-master**, **preset-build-release**, **preset-build-pr**).
- The release job is defined for the `release-0.6` branch.
- The release job has the **always_run** flag set to `true`.

### Define a test for a release job

See an example of a test configuration from the `binding_usage_controller_test.go` file:
```go
func TestBucReleases(t *testing.T) {
  // WHEN
  for _, currentRelease := range tester.GetAllKymaReleaseBranches() {
    t.Run(currentRelease, func(t *testing.T) {
      jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/components/binding-usage-controller/binding-usage-controller.yaml")
      // THEN
      require.NoError(t, err)
      actualPresubmit := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], tester.GetReleaseJobName("kyma-components-binding-usage-controller", currentRelease), currentRelease)
      require.NotNil(t, actualPresubmit)
      assert.False(t, actualPresubmit.SkipReport)
      assert.True(t, actualPresubmit.Decorate)
      assert.Equal(t, "github.com/kyma-project/kyma", actualPresubmit.PathAlias)
      tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, currentRelease)
      tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepo, preset.GcrPush, preset.BuildRelease)
      assert.True(t,actualPresubmit.AlwaysRun)
      tester.AssertThatExecGolangBuidlpack(t, actualPresubmit.JobBase, tester.ImageGolangBuildpackLatest, "/home/prow/go/src/github.com/kyma-project/kyma/components/binding-usage-controller")
    })
  }
}

```

Follow the presented example to implement tests to reuse them in every release.
The example uses the `tester.GetAllKymaReleaseBranches()` function that returns all supported Kyma release branches and runs a separate test for every release branch. If you add a new branch, the tests for the release job are already available. This approach assumes that the job definition does not differ between releases, except for the **branch** and **extra-refs.base_ref** parameters.
