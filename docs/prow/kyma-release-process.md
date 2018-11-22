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

The proposed release process looks as follows in Prow:

1.  Create a release branch in the `kyma` repository. Do it only for a new release, not for a bugfix release.
    The name of this branch should follow the `release-x.y` pattern, such as `release-0.4`.
2.  Create a pull request (PR). Set the **version** and **dir** parameters in `values.yaml` files in root charts to those of the release version.
    For every component, there is a presubmit job for release branches, that is responsible for publishing Docker images with a release tag, such as `0.4.3`.

```
    name: prow/kyma/components/ui-api-layer/release
    branches:
      - release-X.Y
    run_if_changed: ...values.yaml,components-ui-api-layer
```

For details on how to calculate the proper image tag, see the **Calculate the release image tag** [section](#calculate-the-release-image-tag).

Such a configuration ensures that after modifying all `values.yaml` files successfully, all images are published.
There can be an additional job that checks if all components in all `values.yaml` files use exactly the same version, such as `0.4.3`.
You must define branch protection rules for release branches and the `master` branch. These rules mark all checks as a required.
Without them, the **Squash and merge** button is enabled even if some checks failed or are in progress.

3. Run the integration jobs. Use one of these options to configure them:

- Run it on every change on a PR. This is a very simple solution from Prow's perspective. Unfortunately, there is a high possibility that this job fails at the beginning.
  This job should be run only after all
  components are already built and images are already published. Unfortunately, there is no option to configure Prow jobs in that way.
  You need to retrigger failed integration tests by adding a `retest` comment on a PR.
- Run integration tests on demand. A comment on a PR, instead of a change in a source code, triggers the job.
  The Release Manager is responsible for adding such a comment.
- Run integration tests after merging to the release branch. This is a very simple solution, but it can happen that on a release branch we have changes
  that do not pass integration tests.

4. Publish artifacts and create the release. Use one of the options to configure them:

- Run jobs after integrations jobs:

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

You can configure it either as a presubmit job that runs on a PR or a postsubmit job that runs after you merge a PR.
With a presubmit job, it is easy to retrigger such an action by adding a proper command to a PR. Such an option is not available for
a postsubmit job.

- Trigger jobs by specifying comments on a PR. This can be an option, if there are some manual steps required before publishing artifacts.

### Consequences

In the proposed approach, you have to define an additional presubmit job for every component responsible for releasing a given component.
Such a job is almost the same as the presubmit job for the `master` branch, except for the differences in:

- Branch name.
- **run_if_changed** parameters.
- Labels, since you use `preset-build-release` instead of `preset-build-pr`.
  To reduce boilerplate code and code repetition, use `yaml` features, such as extending objects.

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
          - "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"
        args:
          - "/home/prow/go/src/github.com/kyma-project/kyma/components/ui-api-layer"

job_labels_template: &job_labels_template
  preset-dind-enabled: "true"
  preset-sa-gcr-push: "true"
  preset-docker-push-repository: "true"

presubmits: # runs on PRs
  kyma-project/kyma:
    - name: prow/kyma/components/ui-api-layer
      branches:
        - master
      run_if_changed: "components/ui-api-layer/"
      labels:
        preset-build-pr: "true"
        <<: *job_labels_template
      <<: *job_template
    - name: prow/release/kyma/components/ui-api-layer
      run_if_changed: "(components/ui-api-layer/|resources/core/values.yaml)"
      branches:
        - '^release-\d+\.\d+$'
      labels:
        preset-build-release: "true"
        <<: *job_labels_template
      <<: *job_template
```

### Calculate the release image tag

In the internal CI (Jenkins), used for the release process, it is easy to specify which version to publish by adding it as a job parameter.
In Prow, such an option is not available. Instead, you need to calculate this version. This proposal describes how to do it be reading Git tags.
See the proposal for the **export_variables** function in `prow/scripts/library.sh`, which shows how to calculate a new tag:

```bash
function export_variables() {
    if [[ "${BUILD_TYPE}" == "pr" ]]; then
        DOCKER_TAG="PR-${PULL_NUMBER}"
    elif [[ "${BUILD_TYPE}" == "master" ]]; then
        DOCKER_TAG="$(git describe --tags --always)"
     elif [[ "${BUILD_TYPE}" == "release" ]]; then
        echo "Calculating DOCKER_TAG variable for release..."
        branchPattern="^release-[0-9]+\.[0-9]+$"
        echo ${PULL_BASE_REF} | grep -E -q ${branchPattern}
        branchMatchesPattern=$?
        if [ ${branchMatchesPattern} -ne 0 ]
        then
            echo "Branch name does not match pattern: ${branchPattern}"
            exit 1
        fi

        version=${PULL_BASE_REF:8}
        # Getting last tag that matches version
        last=$(git tag --list "${version}.*" --sort "-version:refname" | head -1)

        if [ -z "$last" ]
        then
            newVersion="${version}.0"
        else
            tagPattern="^[0-9]+.[0-9]+.[0-9]+$"
            echo ${last} | grep -E -q ${tagPattern}
            lastTagMatches=$?
            if [ ${lastTagMatches} -ne 0 ]
            then
                echo "Last tag does not match pattern: ${tagPattern}"
                exit 1
            fi

            list=(`echo ${last} | tr '.' ' '`)
            vMajor=${list[0]}
            vMinor=${list[1]}
            vPatch=${list[2]}
            vPatch=$((vPatch + 1))
            newVersion="$vMajor.$vMinor.$vPatch"
        fi
         echo "New version is $newVersion"
         DOCKER_TAG=$newVersion

    else
        echo "Not supported build type - ${BUILD_TYPE}"
        exit 1
    fi
    readonly DOCKER_TAG
    export DOCKER_TAG
}
```
