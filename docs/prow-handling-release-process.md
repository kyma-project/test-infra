# Prow - Handling release process

Purpose of this file is to define how we want to handle release process using Prow.

## Current state
Current release process should be documented in this [issue](https://github.com/kyma-project/community/issues/105).

1. Create release branch in kyma repository - only done when creating new release, not bugfix release
        a. Name of this branch should follow pattern release-x.y (example release-0.4)
2. Bump versions and dirs parameters in values.yaml files in root charts, setting those to desired version which will be released. 
Commit and push this file to release branch.
3. Run release pipeline on Jenkins (`release.Jenkinsfile`). We run release plans for every component (with suffix `-release`). 
  Difference between release plan, and normal plan for component is that in release plan, merging strategy is disabled.  
The merging strategy needs to be disabled for time of running release pipeline). 
4. Run pipeline for releasing kyma-installer from release branch and set version to release version
5. Produce local and cluster combo yamls
        a. ./installation/scripts/generate-kyma-installer.sh ./installation/resources/installer-config-local.yaml.tpl > kyma-config-local.yaml
        b. ./installation/scripts/generate-kyma-installer.sh ./installation/resources/installer-config-cluster.yaml.tpl > kyma-config-cluster.yaml
6. In generated combo yamls set apropriate kyma-installer image (produced in step 4)
7. Attach artifacts and readme to release on github 

## Proposal: release process in prow
1.   Create release branch in kyma repository - only done when creating new release, not bugfix release
    a. Name of this branch should follow pattern release-x.y (example release-0.4) - nothing changed
2. Create PR. Bump versions and dirs parameters in values.yaml files in root charts, setting those to desired version which will be released.
For every component, define Presubmit job:
```
    name: prow/kyma/components/ui-api-layer/release
    branches:
      - release-X.Y
    run_if_changed: ...values.yaml
```
These jobs are responsible for publishing docker images with release tag (eg. 0.4.3)
TODO: check if branches can be specified as a regexp

After successful modification all values.yaml files we will have all images published. Merge PR to release branch.
3. When merging to release branch, Postsubmit job will be triggered:
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
In this solution, we have to define additional Presubmit job for every component, which differs from Presubmit job defined for branch master:
- branch name
- run_if_changed
- docker image tag

This can be solved by yaml references.


