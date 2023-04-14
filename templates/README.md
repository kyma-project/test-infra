# Templates

## Overview

Jobs and Prow configuration are generated from templates by the Render Templates tool. Check the [Render Templates documentation](../development/tools/cmd/rendertemplates/README.md) for details about usage.

The `templates` directory has the following structure:

- `data`, which is the subdirectory with files that describe jobs that the [Render Templates](../development/tools/cmd/rendertemplates) tool should generate using job definitions from templates.
- `templates` which is the subdirectory with all template files that supply the definition of [Prow jobs](../prow/jobs) used in Kyma.
- `config.yaml`, which is the configuration file that describes configuration and jobs that the [Render Templates](../development/tools/cmd/rendertemplates) tool should generate using job definitions from templates.

The template list includes:

- `generic.tmpl`, which is used to create most of the job definitions.
- `kyma-github-release.yaml` that is used for creating the GitHub release after merging the release branch to the `main` branch.
- `prow-config.yaml` that serves to create the main Prow configuration without job definitions.
- `releases.go.tmpl` that contains a set of functions for the release which provide the list of currently supported releases and all supported Kyma release branches.
- `testgrid-default.yaml`, which defines a set of testgrid dashbords.
- `whitesource-periodics.tmpl`, which defines a set of periodic jobs that run a Whitesource scan.

### Configuration file

The `config.yaml` file has two keys:

- `global` with a map of values available for all templates
- `templates` with a list of files to generate

The .yaml files in `data` directory have one key:

- `templates` with a list of files to generate

The `config.yaml` and .yaml files in the `data` directory serve as the input files for the Render Templates. The program generates the jobs based on the definition and templates which are specified in the files. These files define the names of the template file and output file, their location, and configuration referred to in `values`.

See the example of `application-gateway`, in which the `generic.taml` template is used to create the component and test-related YAML files using values defined by the **kyma_generic_component** parameter.

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
                    global:
                      - "jobConfig_presubmit"
                  postConfigs:
                    global:
                      - "jobConfig_postsubmit"
                      - "disable_testgrid"
        - to: ../prow/jobs/kyma/tests/application-gateway-tests/application-gateway-tests-generic.yaml
          localSets:
            jobConfig_pre:
              labels:
                preset-build-pr: "true"
            jobConfig_post:
              labels:
                preset-build-main: "true"
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
                      global:
                        - "jobConfig_presubmit"
                      local:
                        - "jobConfig_pre"
                    postConfigs:
                      global:
                        - "jobConfig_postsubmit"
                        - "disable_testgrid"
                      local:
                        - "jobConfig_post"
```

### Component templates

Component jobs are defined similarly to a regular job, with the exception that the `name` field must be empty (because the name is generated by the Render Templates tool), and the `path` value must be set.

Component job will generate presubmit and postsubmit jobs for the next release, and by default, it will also generate these jobs for supported releases.
The rest of the values is copied from the main jobConfig to the generated ones.

A template receives two objects as input:
- `Values` which contains all the values specified under `values` in the `config.yaml` file.
- `Global` which contains values specified under `global` in the `config.yaml` file.

See the description of values used by component job templates:

| Name | Required | Description |
|------| :-------------: |------|
| **name** | No | Name must not be set. It is generated for each job. |
| **path** | Yes | Path in a repository to the component files. |
| **release_since** | No |  Specifies the release from which this component version applies. |
| **release_since** | No |  Specifies the release till which this component version applies.  |
| **skipReleaseJobs** | No | Specifies if the Render Templates tools should omit generating job definitions for currently supported releases. |

All the functions from the [`sprig`](https://github.com/Masterminds/sprig) library are available in the templates. It is the same library that is used by Helm, so if you know Helm, you are already familiar with them. Also, a few additional functions are available:
- `releaseMatches {release} {since} {until}` returns a boolean value indicating whether `release` fits in the range. Use `nil` to remove one of the bounds. For example, `releaseMatches {{ $rel }} '1.2' '1.5'` checks if the release `$rel` is not earlier than `1.2` and not later than `1.5`.
- `matchingReleases {all-releases} {since} {until}` returns a list of releases filtered to only those that fit in the range.
