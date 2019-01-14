# Release checklist

This document describes how to execute a release of Kyma using prow.

## Release preparation

This part can be done at any point in time. 
This part applies only for new major and minor versions.

1. Define release jobs in test-infra repository
 - for every component
 - for every tool
 - for every test
 - kyma-docs
 - kyma-integration
 - kyma-gke-integration
 - kyma-artifacts
 - kyma-installer

To see all release jobs for release 0.6, look for job names with prefix: `pre-rel06`.
Every job name has to have a unique name, so remember about prefixing it with `pre-rel<XY>`.
Prow reads job configuration from the `master` branch of the `test-infra` repository, so release jobs should also be defined on the `master` branch.

2. Ensure that tests for release jobs exist. Release tests usually iterate through all release versions and run tests for them.
See test `TestBucReleases` defined in `development/tools/jobs/kyma/binding_usage_controller_test.go` as a reference.
To add tests for all jobs for the new release just update function `GetAllKymaReleaseBranches()`
  defined in `development/tools/jobs/tester/tester.go` file from
`test-infra` repository.

3. Define branch protection rules for the release branch in `prow/config.yaml` file.
For release-0.6 definition was the following:
```
release-0.6:
  protect: true
  required_status_checks:
    contexts:
      - pre-rel06-kyma-integration
      - pre-rel06-kyma-gke-integration
      - pre-rel06-kyma-artifacts
      - pre-rel06-kyma-installer
```


## Release
1. Create a release branch in `test-infra` repository.
This part applies only for new major and minor versions.
The name of this branch should follow the `release-x.y` pattern, such as `release-0.6`.

2. Ensure that the `prow/RELEASE_VERSION` file from `test-infra` repository on a release branch contains correct version to be created.
File should contains version if format `<A>.<B>.<C>` or `<A>.<B>.<C>-rc<D>` where `A`,`B`, `C` and `D` are numbers.
If you define release candidate version, pre-release will be created.

3. Create a release branch in the `kyma` repository. Do it only for a new release, not for a bugfix release.
The name of this branch should follow the `release-x.y` pattern, such as `release-0.6`.

4. Create a PR for the `kyma` release branch.

![](./assets/release-PR.png)
This triggers all jobs for components.
In a PR you have to update version and directory of used components in `values.yaml` files.

Change from:

```
dir: develop/
version: <current_version>
```
to:
```
dir:
version: <rel_version>
```    
    
Every component image is published with a version defined in the `RELEASE_VERSION` file stored in the `test-infra` repository on the given release branch. 
5. In case of a temporary failure of any job, you can retrigger it by adding a comment to the PR:
```
/test <job_name>
```

6. Wait until all components for jobs or tools will be finished. 
`kyma-integration`, `kyma-gke-integration`, `kyma-artifacts` and `kyma-installer` needs to be executed manually because there
are dependencies between jobs. See diagram: 
![](./assets/kyma-rel-jobs.svg)

7.  Execute `kyma-integration` by adding comment on PR:
`/test pre-rel06-kyma-integration`

8. Execute one after the other `kyma-installer`, `kyma-artifacts` and `kyma-gke-integration` jobs. 
You don't have to wait until `pre-rel06-kyma-integration` job will be finished.

5. In case of  detecting problems with a release, for example failing tests, you need to wait for the fix (fixed on PR or cherry picked to the PR). 
Then all jobs will be triggered one more time and you need to return to point 6.

9. After all checks pass, merge PR.

10. Merge to the release branch, executes Postsubmit job that creates a Github release.
Validate generated `yamls` and changelog files: https://github.com/kyma-project/kyma/releases
Update manually release content with instruction how to install released Kyma.

11. Update `RELEASE_VERSION` to next version immediately after the release, both on master and release branch. If not, any PR to a release branch done by 
Kyma developer overrides previously published Docker images.  
