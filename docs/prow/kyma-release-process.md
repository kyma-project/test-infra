# Kyma - Release Process MVP

The purpose of this document is to define the Kyma release process that uses Prow.
The first Prow implementation aims to support the same functionalities as those supported by the internal CI.

## Release process in the internal CI

The internal release process is documented as a comment in [this](https://github.com/kyma-project/community/issues/105) issue and looks as follows:

1. Create a release branch in the `kyma` repository. Do it only for a new release, not for a bugfix release.
   The name of this branch should follow the `release-x.y` pattern, such as `release-0.4`.
2. Set the **version** and **dir** parameters in `values.yaml` files in root charts to those of the release version.
   Commit and push this file to the release branch.
3. Run the release pipeline on Jenkins (`release.Jenkinsfile`).The pipeline runs plans for every component with the `-release` suffix.
   The difference between the release plan and the non-release plan for a component is that in the release plan, the merging strategy is disabled. The merging strategy needs to be disabled for the time when the release pipeline is running.
4. Run the pipeline to release the Kyma Installer component from the release branch and set **version** to the release version.
5. Produce combined local and cluster `yaml` files:

   a. `./installation/scripts/generate-kyma-installer.sh ./installation/resources/installer-config-local.yaml.tpl > kyma-config-local.yaml`

   b. `./installation/scripts/generate-kyma-installer.sh ./installation/resources/installer-config-cluster.yaml.tpl > kyma-config-cluster.yaml`

6. Set an appropriate Kyma Installer image in the combined `yaml` files.
7. Attach artifacts and a `README.md` document to the release on GitHub.

## Release process in Prow
When you use standard Kyma buildpacks, such as the buildpack for Go or Node.js applications, a job for a component additionally clones the `test-infra` repository to access build scripts.
To gain reproducibility, you must define separate release jobs for every component in a single release. In the release job, clone the `test-infra` repository from the equivalent release branch. For example, when you define the job for releasing the 0.6 version, clone `test-infra` from the `release-0.6` branch.

Since it is easier to analyze the results of presubmit jobs than those of postsubmit jobs, perform all activities related to releasing a new version in the context of a pull request (PR) issued for the release branch.
When you create a PR, Prow triggers a job for every component.
After introducing all necessary changes, such as modifying `values.yaml` with a new version of components, you must trigger additional required jobs by adding a comment to a PR. These jobs are:
- `kyma-installer`
- The job that creates these release artifacts:
    - `kyma-config-local.yaml`
    - `kyma-config-cluster.yaml`
- `kyma-integration`
- `kyma-gke-integration` which should not build the Kyma Installer but use the already released version instead

You can merge the PR to the release branch only after all checks pass. The merge triggers a postsubmit job that creates a
GitHub release and adds a Git tag to it. In case of a release candidate, a pre-release is created.   

### Action plan for releases

1. Prepare the release.

In this phase, you define release jobs for every component. Ensure that tests for jobs exist and modify branch protection rules for the release branch.
This phase only applies to releasing major or minor versions, when the release branch is created.


When you add release jobs, the configuration file for a sample component looks as follows:
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
      extra_refs:
      - <<: *test_infra_ref
        base_ref: master
      run_if_changed: "^components/binding-usage-controller/"
      labels:
        <<: *job_labels_template
        preset-build-master: "true"
```
The differences between a release job and a job for the `master` branch are as follows:

- Different branches
- The `preset-build-release` label used instead of `preset-build-pr`
- The **extra_refs** parameter for the `test-infra` repository that uses the `release-0.6` branch instead of `master`
- The **always_run** parameter set to `true` instead of specifying the **run_if_changed** parameter

Creating a new release branch for every release adds approximately 10 lines for every component:
```
  - branches:
      - release-0.7
      <<: *job_template
      always_run: true
      extra_refs:
      - <<: *test_infra_ref
        base_ref: release-0.7
      labels:
        <<: *job_labels_template
        preset-build-release: "true"
```

If you define a new job, you must define a test for it.

```go
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
			tester.AssertThatHasExtraRefToTestInfraBranch(t, actualPresubmit.JobBase.UtilityConfig, currentRelease)
			tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildRelease)
			assert.True(t,actualPresubmit.AlwaysRun)
			tester.AssertThatExecGolangBuidlpack(t, actualPresubmit.JobBase, tester.ImageGolangBuildpackLatest, "/home/prow/go/src/github.com/kyma-project/kyma/components/binding-usage-controller")
		})
	}
}
```
The example shows that the job's test runs for a given component for every release. See the iteration demonstrated through an array defined as `tester.GetAllKymaReleaseBranches()`.
If there are no changes in Prow job definitions between releases, modify only the `GetAllKymaReleaseBranches` function in the tests.

As the last step in this phase, modify the branch protection rules.
Because some jobs are triggered manually, mark them explicitly as required under `branch-protection` in the `prow/config.yaml` configuration file.
```
    release-0.6:
      required_status_checks:
        contexts:
          - kyma-integration
          - kyma-gke-integration
          - ...
```

2. Create a release branch in the `test-infra` repository.

This step ensures builds reproducibility. It only applies to major and minor releases.

3. Define the release version.

Release candidates are usually created before a final version of the release. The `RELEASE_VERSION` file in the `test-infra` repository contains the information on the released version.

4. Create a release branch in the `kyma` repository. Do it only for a new release, not for a bugfix release.
The name of this branch should follow the `release-x.y` pattern, such as `release-0.4`.

5. Create a PR for the `kyma` release branch. This triggers all jobs for components.
Every component image is published with a version defined in the `RELEASE_VERSION` file stored in the `test-infra` repository on the given release branch. For example, it can include the `0.6.0-rc1` version.

6. Run integration tests by adding a comment to a PR:
```
/test kyma-integration
```
```
/test kyma-gke-integration

```

7. If all checks pass, merge the PR and check if the postsubmit job succeeds. If the job finishes with an error, retrigger it from the Prow dashboard available at `https://status.build.kyma-project.io/`.

### Calculate the release image tag

In the internal CI (Jenkins) used for the release process, it is easy to specify which version to publish by adding it as a job parameter.
In Prow, such an option is not available. Instead, you read that information from the `RELEASE_VERSION` file defined in the `test-infra` repository.

### Remove components
Do not modify release jobs for previous versions of Kyma components that you remove or rename.