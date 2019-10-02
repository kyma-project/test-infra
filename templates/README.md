# Templates

## Overview

Jobs and Prow configuration are generated from templates. The `templates` directory contains these items:

- `templates` - the directory with all template files that supply the definition of [Prow jobs](https://github.com/kyma-project/test-infra/tree/master/prow/jobs) used in Kyma.
- `config.yaml` - the configuration file that describes jobs that are to be generated using templates.

The template list includes:

- `compass-integration.yaml` - defines presubmit and postsubmit integration jobs that build Kyma to check its compatibility with the Compass component
- `component.yaml` - defines component jobs with specified buildpacks.
- `generic-component.yaml` - provides a new way of creating component jobs. Instead of a predefined buildpack, it uses a generic bootstrap that contains Makefile and Docker and no component-specific dependencies.
- `kyma-artifacts.yaml`- for creating release artifacts
- `kyma-github-release.yaml` - for creating the Github release after merging the release branch to the master branch
- `kyma-installer.yaml` - fr building the installer used during the release
- `kyma-integration.yaml` - a set of presubmit and postsubmit integration jobs that build Kyma on clusters to verify if the introduced changes did not affect other Kyma components
- `kyma-release-candidate.yaml` - building the release cluster for testing purposes
- `prow-config.yaml` - serves for creating the main Prow configuration without job definitions.
- `releases.go.tmpl` - with a list of functions for the release that provide the list of currently supported releases and all supported Kyma release branches

### Configuration file

The `config.yaml` file has two keys:

- `global` - a map of values available for all templates
- `templates` - a list of files to generate

The `config.yaml` serves as the input file for the [Render Templates](../development/tools/cmd/rendertemplates) tool that generates the jobs based on the file definition and templates which it specifies. The `config.yaml` defines the names of the output file, their location, and configuration referred to in `values`.

See the example of `console-backend-service` in which the `generic-component.yaml` template is used to create the component and test-related yaml files using values defined by the `kyma_generic_component` parameter.

```yaml
kyma_generic_component: &kyma_generic_component
  repository: github.com/kyma-project/kyma
  pushRepository: kyma
  bootstrapTag: v20181204-a6e79be
  additionalRunIfChanged:
    - ^scripts/

templates:
  - from: templates/generic-component.yaml
    render:
      - to: ../prow/jobs/kyma/components/console-backend-service/console-backend-service-generic.yaml
        values:
          <<: *kyma_generic_component
          path: components/console-backend-service
          since: "1.6"
      - to: ../prow/jobs/kyma/tests/console-backend-service/console-backend-service-tests-generic.yaml
        values:
          <<: *kyma_generic_component
          path: tests/console-backend-service
          since: "1.6"

```

### Component templates

A template receives two objects as input:
- `Values` which contains all the values specified under `values` in the configuration file.
- `Global` which contains values specified under `global` in the configuration file.

There are two templates that contain the default configuration for the Prow component jobs, `component.yaml` and `generic-component.yaml`. The recommended template is `generic-component.yaml` that is the new version of the `component.yaml` template that uses a generic bootstrap to build components, instead of various buildpacks.

See the description of values used by both component templates:

| Name | Required | Component template(s) | Description |
| ---| : --- : | ---  | --- |
| `additionalRunIfChanged` | Yes | `generic-component.yaml` | Provides a list of regexps for Prow to watch in addition to `path`. Prow runs the job if it notices any changes in the specified files or folders. The default value is `[]`. |
| `bootstrapTag` | Yes | `generic-component.yaml` | Provides the tag of the bootstrap image to use. |
| `buildpack` | Yes | `component.yaml` | Specifies the buildpack version used to build the component. |
| `env` | No | `component.yaml` | Specifies the environment variable that turns on Go modules required to build Kubebuilder v2. |
| `noRelease` | No | `component.yaml` | Specifies that this component does not require a release job. |
| `optional` | No | Both | Defines if this job is obligatory or optional on pull requests. Set it to `true` when you add a new component and remove it after the whole CI pipeline for the component is in place. |
| `patchReleases` | No | `component.yaml` | _TODO_ |
| `path` | Yes | Both | Specifies the location of the component in the repository, such as `components/console-backend-service`. |
| `presets.build` | Yes | `component.yaml` | _TODO_ |
| `presets.pushRepository` | Yes | `component.yaml` | _TODO_ |
| `pushRepository` | Yes | `generic-component.yaml` | Provides the suffix of the `preset-docker-push-` label to define the GCR image location, such as `kyma`. |
| `ReleaseBranchPattern` | Yes | `generic-component.yaml` | Defines the prefix pattern of the release branch for which Prow should run the release job. The default value is `^release-{supported-releases}-{component-dir-name}$`. |
| `repository` | Yes | Both | Specifies the component's GitHub repository address, such as `github.com/kyma-project/kyma`. |
| `resources.memory` | Yes | Both | Specifies the memory assigned to the job container. The default value is `1.5Gi`. |
| `resources.cpu` | Yes | Both | Specifies the CPU assigned to the job container. The default value is `0.8`. |
| `since` | Yes | Both | Specifies the release from which the component is active. |
| `until` | Yes | Both | Specifies the release till which the component is active.  |
| `watch` | No | `component.yaml` | Provides a list of regexps for Prow to watch in addition to `path`. Prow runs the job if it notices any changes in the specified files or folders. |


All the functions from [`sprig`](https://github.com/Masterminds/sprig) library are available in the templates. It is the same library that is used by helm, so if you know helm, you are already familiar with them. Also, a few additional functions are available:
- `releaseMatches <release> <since> <until>` returns a boolean value indicating whether `release` fits in the range. Use `nil` to remove one of the bounds. For example: `releaseMatches {{ $rel }} '1.2' '1.5'` checks if the release `$rel` is younger than `1.2` and older than `1.5`.
- `matchingReleases <all-releases> <since> <until>` returns a list of releases filtered to only those that fit in the range.
