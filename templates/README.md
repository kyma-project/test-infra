# Templates

## Overview

Jobs and Prow configuration are generated from templates. The `templates` directory has the following structure:

- `templates` which is the subdirectory with all template files that supply the definition of [Prow jobs](../prow/jobs) used in Kyma.
- `config.yaml` which is the configuration file that describes jobs that the [Render Templates](../development/tools/cmd/rendertemplates) tool should generate using job definitions from templates.

The template list includes:

- `branchprotector-config.yaml` that defines configuration for branch protection. Provide any changes related to branch protection in that file.
- `compass-integration.yaml` that defines presubmit and postsubmit integration jobs which build Kyma to check its compatibility with the Compass component.
- `component.yaml` that defines component jobs with specified buildpacks.
- `generic-component.yaml` that provides a new way of creating component jobs. Instead of a predefined buildpack, it uses a generic bootstrap that contains Makefile and Docker but no component-specific dependencies.
- `kyma-artifacts.yaml` that serves to create release artifacts.
- `kyma-github-release.yaml` that is used for creating the GitHub release after merging the release branch to the `master` branch.
- `kyma-installer.yaml` that is used for building the Installer during the release.
- `kyma-integration.yaml` that defines a set of presubmit and postsubmit integration jobs that build Kyma on clusters to verify if the introduced changes do not affect the existing Kyma components.
- `kyma-release-candidate.yaml` that is used for building the release cluster for testing purposes.
- `prow-config.yaml` that serves to create the main Prow configuration without job definitions.
- `releases.go.tmpl` that contains a set of functions for the release which provide the list of currently supported releases and all supported Kyma release branches.

### Configuration file

The `config.yaml` file has two keys:

- `global` with a map of values available for all templates
- `templates` with a list of files to generate

The `config.yaml` serves as the input file for the Render Templates that generates the jobs based on the file definition and templates which it specifies. The `config.yaml` file defines the names of the output file, their location, and configuration referred to in `values`.

See the example of `console-backend-service` in which the `generic-component.yaml` template is used to create the component and test-related YAML files using values defined by the **kyma_generic_component** parameter.

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
- `Values` which contains all the values specified under `values` in the `config.yaml` file.
- `Global` which contains values specified under `global` in the `config.yaml` file.

There are two templates that contain the default configuration for the Prow component jobs, `component.yaml` and `generic-component.yaml`. The recommended template is `generic-component.yaml` that is the new version of the `component.yaml` template and uses a generic bootstrap to build components, instead of various buildpacks.

See the description of values used by both component templates:

| Name | Required | Component template(s) | Description |
|------| :-------------: |------| ------|
| **additionalRunIfChanged** | No | `generic-component.yaml` | Provides a list of regexps for Prow to watch in addition to `path`. Prow runs the job if it notices any changes in the specified files or folders. The default value is `[]`. |
| **bootstrapTag** | Yes | `generic-component.yaml` | Provides the tag of the bootstrap image to use. |
| **buildpack** | Yes | `component.yaml` | Specifies the buildpack version used to build the component. |
| **env** | No | `component.yaml` | Specifies the environment variable that turns on Go modules required to build Kubebuilder v2. |
| **noRelease** | No | `component.yaml` | Specifies that this component does not require a release job. |
| **optional** | No | Both | Defines if this job is obligatory or optional on pull requests. Set it to `true` when you add a new component and remove it after the whole CI pipeline for the component is in place. |
| **patchReleases** | No | `component.yaml` | A list of releases that patch the given component version. |
| **path** | Yes | Both | Specifies the location of the component in the repository, such as `components/console-backend-service`. |
| **presets.build** | Yes | `component.yaml` | The name of the Preset for building the component on the `master` branch. For example, set to `build` to use the **preset-build-master** Preset. |
| **presets.pushRepository** | Yes | `component.yaml` | Provides the suffix of the **preset-docker-push-** label to define the GCR image location, such as `kyma`.  |
| **pushRepository** | Yes | `generic-component.yaml` | Provides the suffix of the **preset-docker-push-** label to define the GCR image location, such as `kyma`. |
| **ReleaseBranchPattern** | No | `generic-component.yaml` | Defines the prefix pattern for the release branch for which Prow should run the release job. The default value is `^release-{supported-releases}-{component-dir-name}$`. |
| **repository** | Yes | Both | Specifies the component's GitHub repository address, such as `github.com/kyma-project/kyma`. |
| **resources.memory** | No | Both | Specifies the memory assigned to the job container. The default value is `1.5Gi`. |
| **resources.cpu** | No | Both | Specifies the CPU assigned to the job container. The default value is `0.8`. |
| **since** | Yes | Both | Specifies the release from which this component version applies. |
| **until** | Yes | Both | Specifies the release till which this component version applies.  |
| **watch** | No | `component.yaml` | Provides a list of regexps for Prow to watch in addition to `path`. Prow runs the job if it notices any changes in the specified files or folders. |


All the functions from the [`sprig`](https://github.com/Masterminds/sprig) library are available in the templates. It is the same library that is used by Helm, so if you know Helm, you are already familiar with them. Also, a few additional functions are available:
- `releaseMatches {release} {since} {until}` returns a boolean value indicating whether `release` fits in the range. Use `nil` to remove one of the bounds. For example, `releaseMatches {{ $rel }} '1.2' '1.5'` checks if the release `$rel` is not earlier than `1.2` and not later than `1.5`.
- `matchingReleases {all-releases} {since} {until}` returns a list of releases filtered to only those that fit in the range.
