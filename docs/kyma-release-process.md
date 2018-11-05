# Kyma - Release Process MVP
Purpose of this document is to define how we want to support release process for Kyma using Prow.
In the first implementation, we want to support the same functionality as it done in internal CI. 

## Current state
Current release process is documented as a comment in [this issue](https://github.com/kyma-project/community/issues/105).

>1. Create release branch in kyma repository - only done when creating new release, not bugfix release
        a. Name of this branch should follow pattern release-x.y (example release-0.4)
>2. Bump versions and dirs parameters in values.yaml files in root charts, setting those to desired version which will be released. 
Commit and push this file to release branch.
>3. Run release pipeline on Jenkins (`release.Jenkinsfile`). We run release plans for every component (with suffix `-release`). 
  Difference between release plan, and normal plan for component is that in release plan, merging strategy is disabled.  
The merging strategy needs to be disabled for time of running release pipeline). 
>4. Run pipeline for releasing kyma-installer from release branch and set version to release version
>5. Produce local and cluster combo yamls
>
>   a. ./installation/scripts/generate-kyma-installer.sh ./installation/resources/installer-config-local.yaml.tpl > kyma-config-local.yaml
>
>   b. ./installation/scripts/generate-kyma-installer.sh ./installation/resources/installer-config-cluster.yaml.tpl > kyma-config-cluster.yaml
>6. In generated combo yamls set apropriate kyma-installer image (produced in step 4)
>7. Attach artifacts and readme to release on github 

## Proposal: release process in Prow
1.   Create release branch in kyma repository - only done when creating new release, not bugfix release.
    a. Name of this branch should follow pattern release-x.y (example release-0.4)
2. Create PR. Bump versions and dirs parameters in values.yaml files in root charts, setting those to desired version which will be released.
For every component, there is a presubmit job for release branches, that is responsible for publishing docker images with release tag (eg. 0.4.3). 
```
    name: prow/kyma/components/ui-api-layer/release
    branches:
      - release-X.Y
    run_if_changed: ...values.yaml,components-ui-api-layer
```
Calculation of proper image tag is described in section "Calculation release image tag" [here](#calculation-release-image-tag)

Such configuration implies that after successful modification of all values.yaml files, all images are published.
There can be an additional job, that checks if in all `values.yaml` all components use exactly the same version (eq. 0.4.3).
For release branches and master branch, branch protection rules needs to be defined, that mark all checks as a required. Without that, merge button is enabled
even if some checks failed or are in progress. 


3. Integration jobs needs to be executed. There are following options how to configure them:
- run it on every change on PR. This is very simple solution from Prow perspective. Unfortunately, there is a high possibility that this job will fail at the beginning. 
 This job should be executed only after all all 
components are already built and images are already published, but there is no option to configure Prow jobs in that way. Failed integration tests needs to be retriggered, by adding comment `retest` on PR. 
- run integration tests "on demand". Job is not triggered by any change in the source code, but rather is triggered by comment on PR. 
Release Manager is responsible for adding such comment. 
- run integration tests after merging to release branch. This is very simple solution, but it can happen that on a release branch we have changes
that do not pass integration tests.

4. Publishing artifacts and release creation. There are following options how to configure them:
- run jobs after integrations jobs:
```
    name: prow/kyma/integration
    branches:
      - release-X.Y
    // run integration tests
    run_after_success:
      - build kyma installer
      - publish artifacts
      - create release
```
 This can be configured as a Presubmit (on PR) or Postsubmit job. In case of Presubmit job, it is easiliy to retrigger such action
 by adding proper command in PR. In case of Postsubmit job, such option does not exist. 
 - trigger jobs by specifying comments on PR. This can be an option, if there are some manual steps required before publishing artifacts. 

### Consequences
In proposed approach, we have to **define additional Presubmit job for every component responsible for releasing component.** 
Such job is almost the same as Presubmit job for master branch, except:
- branch name
- run_if_changed parameters
- different labels: preset-build-release instead of preset-build-pr
To reduce boilerplate code and code repetition, we can use yaml features, such as extending objects.
```yaml
job_template: &job_template
  optional: false 
  skip_report: false 
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
    context: *name
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

### Calculation release image tag
In previous solution, we used Jenkins for release process. In Jenkins, we can easily specify which version we want to publish, as a job parameter.
In Prow, we don't have such option so we are forced to calculate this version. In this proposal, it is achieved by reading git tags. 
See proposal for export_variables function in prow/scripts/library.sh, where we calculate a new tag.
```bash
function export_variables() {
    if [[ "${BUILD_TYPE}" == "pr" ]]; then
        DOCKER_TAG="PR-${PULL_NUMBER}"
    elif [[ "${BUILD_TYPE}" == "master" ]]; then
        DOCKER_TAG="$(git describe --tags --always)"
     elif [[ "${BUILD_TYPE}" == "release" ]]; then # this is provided by  preset-build-release
        # if we make a PR that wants to merge to branch release-0.4, then PULL_BASE_REF=release-0.4, branchVersion=0.4
        branchVersion=${PULL_BASE_REF:8} 
        # get last tag that was released from current branch. For example we have following tags: 0.4.1, 0.4.2, 0.5.1 and PULL_BASE_REF=release-0.4. 
        # Last variable will be set to 0.4.2 (0.5.1 is ignored)
        last=$(git tag --list "${branchVersion}.*" --sort "-version:refname" | head -1)
        # TODO: handle situation that there is no tag for given branch. 
        list=(`echo $last | tr '.' ' '`)
        vMajor=${list[0]}
        vMinor=${list[1]}
        vPatch=${list[2]}
        vPatch=$((vPatch + 1))

        newVersion="$vMajor.$vMinor.$vPatch" # 0.4.3
        echo "new version is $newVersion "
        DOCKER_TAG=$newVersion
    else
        echo "Not supported build type - ${BUILD_TYPE}"
        exit 1
    fi
    readonly DOCKER_TAG
    export DOCKER_TAG
}
```
