# Kyma - Release Process MVP

The purpose of this document is to define the release process in Kyma using Prow.
The first Prow implementation aims to support the same functionalities as those supported by the internal CI.

## Release process in the internal CI

The internal release process is documented as a comment in [this](https://github.com/kyma-project/community/issues/105) issue and looks as follows:

1. Create a release branch in the `kyma` repository. Do it only for a new release, not for a bugfix release.
   The name of this branch should follow the `release-x.y` pattern, such as `release-0.4`.
2. Set the **version** and **dir** parameters in `values.yaml` files in root charts to those of the release version.
   Commit and push this file to the release branch.
3. Run the release pipeline on Jenkins (`release.Jenkinsfile`).The pipeline runs plans for every component with the `-release` suffix.
   The difference between the release plan and the non-release plan for a component is that in the release plan, the merging strategy is disabled. The merging strategy needs to be disabled for the time when the release pipeline is running.
4. Run the pipeline to release the Kyma-Installer component from the release branch and set **version** to the release version.
5. Produce combined local and cluster `yaml` files:

   a. `./installation/scripts/generate-kyma-installer.sh ./installation/resources/installer-config-local.yaml.tpl > kyma-config-local.yaml`

   b. `./installation/scripts/generate-kyma-installer.sh ./installation/resources/installer-config-cluster.yaml.tpl > kyma-config-cluster.yaml`

6. Set an appropriate Kyma-Installer image in the combined `yaml` files.
7. Attach artifacts and a `README.md` document to the release on GitHub.

## Release process in Prow
When using standard kyma buildpacks, like buildpack for go or node applications, job for a component additionally clones `test-infra` repository to access build scripts.
To gain reproducibility, we have to define separate release jobs for every component. In the release job, cloning of `test-infra` repository should be done from equivalent branch, 
i.e. when defining job for releasing version 0.6, test-infra should be cloned also from branch `release-0.6`.

Considering that Presubmit jobs are more powerful than postsubmits, all activities around releasing new version should be done in the context of a pull request to a release branch.
When creating PR, job for every component will be triggered. 
After intruducing all necessary changes, like modification of `values.yaml` with new version of components, release master 
has to trigger additional, required jobs by adding comment to PR. Those jobs are:
- kyma-installer
- job that create release artifacts
    - kyma-config-local.yaml
    - kyma-config-cluster.yaml
- kyma-integration
- kyma-gke-integration-release (does not build kyma installer, use already released version)


Only after all checks passed, pull request can be merged to a release branch. This will trigger postsubmit job that is responsible for creating 
github release and adding tag. In case of release candidate, prerelease will be created.   

### Action plan for releasing

1. Release preparation
In this phase, we define release jobs for EVERY component, ensure that tests for jobs exists and modify branch protection rules.
This phase needs to be done only for releasing major or minor versions (when release branch is created). 

When adding release jobs, configuration file for sample component looks as follows:
```
test_infra_ref: &test_infra_ref
  org: aszecowka
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
  aszecowka/kyma:
    - name: kyma-components-binding-usage-controller
      branches:
        - &branch master
      <<: *job_template
      extra_refs:
      - <<: *test_infra_ref
        base_ref: *branch
      labels:
        <<: *job_labels_template
        preset-build-master: "true"
```
The difference between a releasing job and Presubmit job for master branch are following:

- branch
- used label `preset-build-release` instead of `preset-build-pr`
- extra refs that clones `test-infra` repository use branch `release-0.6` instead of `master`
- `always-run` set to `true` instead specifying `run_if_changed`

In the next releases (when a new release branch is created), ca 10 lines for every component will be added:
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

When we define a new job, we should also define test for it. 

```go
func TestBucReleases(t *testing.T) {
	// WHEN
	for _, currentRelease := range tester.GetAllKymaReleaseBranches() {
		t.Run(currentRelease, func(t *testing.T) {
			jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/components/binding-usage-controller/binding-usage-controller.yaml")
			// THEN
			require.NoError(t, err)
			actualPresubmit := tester.FindPresubmitJobByName(jobConfig.Presubmits["aszecowka/kyma"], "kyma-components-binding-usage-controller", currentRelease)
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
In the example above, we run job's test for given component for every release (see `tester.GetAllKymaReleaseBranches()`).
If there will be no changes in Prow job definitions between releases, modification of tests is not needed (except global modification of the `GetAllKymaReleaseBranches` function).

Last thing that needs to be done in this phase is to modify branch protection rules.
Because some jobs are triggered manually, we need to mark them explicitly as required in branch-protection configuration file (`prow/config.yaml`).
```
    release-0.6:
      required_status_checks:
        contexts:
          - kyma-integration 
          - kyma-gke-integration
          - ...
```

2. Create release branch in `test-infra` repository.
This is done to ensure builds reproducibility. This phase applies only for major and minor releases.

3. Define release version. 
Usually, we produce some release candidates at the beginning and then final version. This information is stored in file `RELEASE_VERSION` in `test-infra` repository.

3.  Create a release branch in the `kyma` repository. Do it only for a new release, not for a bugfix release.
The name of this branch should follow the `release-x.y` pattern, such as `release-0.4`.
    
4. Create PR (pull request) to `kyma` release branch. This triggers all jobs for components. Every job will be published with version defined in file `RELEASE_VERSION` stored in `test-infra` repository on given release branch, for example `0.6.0-rc1`. 

3. Run integration tests by adding comment to PR:
```
/test kyma-integration
```
```
/test kyma-gke-integration

```


4. If all checks passed, merge PR and observe if postsubmit job succeeded. In case of error, retrigger job from Prow dashboard available at: `https://status.build.kyma-project.io/`


### Calculate the release image tag

In the internal CI (Jenkins), used for the release process, it is easy to specify which version to publish by adding it as a job parameter.
In Prow, such an option is not available. Instead, we read that information from file `RELEASE_VERSION` defined in `test-infra` repository.


### Removing components
Beaware, that in case of removing kyma component or it's renaming, releasing jobs for previous versions should be not modified.