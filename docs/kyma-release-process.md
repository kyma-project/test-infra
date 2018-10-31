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
For every component, we have defined Presubmit jobs:
```
    name: prow/kyma/components/ui-api-layer/release
    branches:
      - release-X.Y
    run_if_changed: ...values.yaml,components-ui-api-layer
```
These jobs are responsible for publishing docker images with release tag (eg. 0.4.3).  
Calculation of proper Docker image tag is described in section "Calcuclation proper image tag"
These jobs in most cases will be the same as presubmit jobs that handles PR for master branch, except:
- branch name
- run_if_changed parameters
- different labels: preset-build-release instead of preset-build-pr


After successful modification of all values.yaml files we will have all images published. Merge PR to release branch.
To discuss: we can define additional Presubmit job which checks if all components have specified proper version. 

3. When merging to release branch, Postsubmit job with integration test will be triggered:
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
Alternatively, it can be Presubmit job run on the PR, that is triggered manually by providing comment on PR (we need to define branch protection that this build is required).  

### Consequences
In this solution, we have to define additional Presubmit job for every component responsible for releasing component. 
To reduce boilerplate code and code repetition, we can yaml features, such as extending objects.

## Investigations
### AAA
TODO: check if branches can be specified as a regexp
### Calcuclation proper image tag

TODO: how to provide proper tag for component
- get env PULL_BASE_REF (release 0.4)
- get highest tag, that has pattern 0.4.X
- current tag will be: 0.4.X+1


see prow/scripts/library.sh
```bash
function export_variables() {
    if [[ "${BUILD_TYPE}" == "pr" ]]; then
        DOCKER_TAG="PR-${PULL_NUMBER}"
    elif [[ "${BUILD_TYPE}" == "master" ]]; then
        DOCKER_TAG="$(git describe --tags --always)"
     elif [[ "${BUILD_TYPE}" == "release" ]]; then
        // put this logic here
    else
        echo "Not supported build type - ${BUILD_TYPE}"
        exit 1
    fi
    readonly DOCKER_TAG
    export DOCKER_TAG
}
```
