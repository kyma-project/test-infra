# Templates

## Overview

Jobs and Prow configuration are generated from templates. The `templates` directory has the following structure:

- `templates` which is the subdirectory with all template files that supply the definition of [Prow jobs](../prow/jobs) used in Kyma.
- `config.yaml` which is the configuration file that describes jobs that the [Render Templates](../development/tools/cmd/rendertemplates) tool should generate using job definitions from templates.

The template list includes:

- `branchprotector-config.yaml` that defines configuration for branch protection. Provide any changes related to branch protection in that file.
- `compass-integration.yaml` that defines presubmit and postsubmit integration jobs which build Kyma to check its compatibility with the Compass component.
- `generic-component.yaml` that provides a new way of creating component jobs. Instead of a predefined buildpack, it uses a generic bootstrap that contains Makefile and Docker but no component-specific dependencies.
- `kyma-artifacts.yaml` that serves to create release artifacts.
- `kyma-github-release.yaml` that is used for creating the GitHub release after merging the release branch to the `main` branch.
- `kyma-integration.yaml` that defines a set of presubmit and postsubmit integration jobs that build Kyma on clusters to verify if the introduced changes do not affect the existing Kyma components.
- `kyma-release-candidate.yaml` that is used for building the release cluster for testing purposes.
- `prow-config.yaml` that serves to create the main Prow configuration without job definitions.
- `releases.go.tmpl` that contains a set of functions for the release which provide the list of currently supported releases and all supported Kyma release branches.

Jobs and Prow configurations have unit tests that are located [here](../development/tools/jobs).

### Configuration file

The `config.yaml` file has two keys:

- `global` with a map of values available for all templates
- `templates` with a list of files to generate

The `config.yaml` serves as the input file for the Render Templates that generates the jobs based on the file definition and templates which it specifies. The `config.yaml` file defines the names of the output file, their location, and configuration referred to in `values`.

See the example of `application-gateway` in which the `generic.taml` template is used to create the component and test-related YAML files using values defined by the **kyma_generic_component** parameter.

```yaml
templates:
  - from: templates/generic.tmpl
    render:
      - to: ../prow/jobs/kyma/components/application-gateway/application-gateway-generic.yaml
        jobConfigs:
          - repoName: "github.com/kyma-project/kyma"
            jobs:
              - jobConfig:
                  path: components/application-gateway
                  args:
                    - "/home/prow/go/src/github.com/kyma-project/kyma/components/application-gateway"
                  run_if_changed: "^components/application-gateway/|^common/makefiles/"
                  release_since: "1.7"
                inheritedConfigs:
                  global:
                    - "jobConfig_default"
                    - "image_buildpack-golang"
                    - "jobConfig_generic_component"
                    - "jobConfig_generic_component_kyma"
                    - "extra_refs_test-infra"
                  preConfigs:
                    - "jobConfig_presubmit"
                  postConfigs:
                    - "jobConfig_postsubmit"
                    - "disable_testgrid"
        - to: ../prow/jobs/kyma/tests/application-gateway-tests/application-gateway-tests-generic.yaml
        jobConfigs:
          - repoName: "github.com/kyma-project/kyma"
            jobs:
              - jobConfig:
                  path: tests/application-gateway-tests
                  args:
                    - "/home/prow/go/src/github.com/kyma-project/kyma/tests/application-gateway-tests"
                  run_if_changed: "^tests/application-gateway-tests/|^common/makefiles/"
                  release_since: "1.7"
                inheritedConfigs:
                  global:
                    - "jobConfig_default"
                    - "image_buildpack-golang"
                    - "jobConfig_generic_component"
                    - "jobConfig_generic_component_kyma"
                    - "extra_refs_test-infra"
                  preConfigs:
                    - "jobConfig_presubmit"
                  postConfigs:
                    - "jobConfig_postsubmit"
                    - "disable_testgrid"
```

### Component jobs

Component jobs are defined similarly to a regular job, with the exception that `name` field has to be empty, as it will be generated; and `path` value has to be set.
Component job will generate presubmit and postsubmit jobs for the next release and by default, it will also generate these jobs for supported releases.
The rest of the values will be copied from the main jobConfig to the generated ones.

See the description of values used by component jobs:

| Name | Required | Description |
|------| :-------------: |------|
| **name** | No | Name must not be set, as it will be generated for each job. |
| **path** | Yes | Path in a repository to the component. |
| **release_since** | No |  Specifies the release from which this component version applies. |
| **release_since** | No |  Specifies the release till which this component version applies.  |

A template receives two objects as input:
- `Values` which contains all the values specified under `values` in the `config.yaml` file.
- `Global` which contains values specified under `global` in the `config.yaml` file.

All the functions from the [`sprig`](https://github.com/Masterminds/sprig) library are available in the templates. It is the same library that is used by Helm, so if you know Helm, you are already familiar with them. Also, a few additional functions are available:
- `releaseMatches {release} {since} {until}` returns a boolean value indicating whether `release` fits in the range. Use `nil` to remove one of the bounds. For example, `releaseMatches {{ $rel }} '1.2' '1.5'` checks if the release `$rel` is not earlier than `1.2` and not later than `1.5`.
- `matchingReleases {all-releases} {since} {until}` returns a list of releases filtered to only those that fit in the range.
