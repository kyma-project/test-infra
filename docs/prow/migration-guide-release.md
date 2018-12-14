# Release Jobs Migration Guide

This document describe the procedure of defining release Jobs for kyma components. There is an assumption that reader is 
familiar with migration guide for standard jobs available [here](https://github.com/kyma-project/test-infra/blob/master/docs/prow/migration-guide.md).

## Steps

### Implement release rule
In your component's Makefile, please make sure that you defined released rule `ci-release`.
Example from Binding Usage Controller:
```
APP_NAME = binding-usage-controller
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG = $(DOCKER_TAG)
binary=$(APP_NAME)

...

.PHONY: ci-release
ci-release: build build-image push-image

```

### Define release job
The difference between a releasing job and Presubmit job for master branch are following:

- branch
- used label `preset-build-release` instead of `preset-build-pr`
- extra refs that clones `test-infra` repository use branch `release-0.6` instead of `master`
- `always-run` set to `true` instead specifying `run_if_changed`


Full example:
```
test_infra_ref: &test_infra_ref
  org: kyma-project
  repo: test-infra
  path_alias: github.com/kyma-project/test-infra

job_template: &job_template
  name: kyma-components-binding-usage-controller
  skip_report: true
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

job_labels_template: &job_labels_template
  preset-dind-enabled: "true"
  preset-sa-gcr-push: "true"
  preset-docker-push-repository: "true"

presubmits: # runs on PRs
  kyma-project/kyma:
  - branches:
    - master
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
    - master
    <<: *job_template
    run_if_changed: "^components/binding-usage-controller/"
    extra_refs:
    - <<: *test_infra_ref
      base_ref: maser
    labels:
      <<: *job_labels_template
      preset-build-master: "true"

```

Please note that:

-`test-infra-ref` object was defined, where `org`, `repo` and `path_alias` is defined.
- `job-template` now defines `name`, but `run_if_changed` and `extra-refs` were removed from it. 
`run_if_changed` is defined only for Presubmit and Postsubmit job for `master` branch.
- all jobs has to define proper `extra-refs` with specified `base-ref`
- every job use different build preset (`preset-build-master`, `preset-build-release`, `preset-build-pr`).
- releasing job is defined for branch `release-0.6`
- release job has `always_run` flag set to `true`

### Define test for release jobs

See example from `binding_usage_controller_test.go`:
```
func TestBucReleases(t *testing.T) {
	// WHEN
	for _, currentRelease := range tester.GetAllKymaReleaseBranches() {
		t.Run(currentRelease, func(t *testing.T) {
			jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/components/binding-usage-controller/binding-usage-controller.yaml")
			// THEN
			require.NoError(t, err)
			actualPresubmit := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], "kyma-components-binding-usage-controller", currentRelease)
			require.NotNil(t, actualPresubmit)
			assert.True(t, actualPresubmit.SkipReport)
			assert.True(t, actualPresubmit.Decorate)
			assert.Equal(t, "github.com/kyma-project/kyma", actualPresubmit.PathAlias)
			tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, currentRelease)
			tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildRelease)
			assert.True(t,actualPresubmit.AlwaysRun)
			tester.AssertThatExecGolangBuidlpack(t, actualPresubmit.JobBase, tester.ImageGolangBuildpackLatest, "/home/prow/go/src/github.com/kyma-project/kyma/components/binding-usage-controller")
		})
	}
}

```

Because we have to be prepared for supporting many releases and for every release separate job for every component needs to be defined, we suggest to implement tests 
as presented above.
In example above, we use `tester.GetAllKymaReleaseBranches()` function that returns all supported Kyma release branches and run separete test for every release branch.
If new branch will be added, we have automatically test for release jobs. In this approach we assume that job definition does not differ between releases except `branch` and `extra-refs.base_ref`