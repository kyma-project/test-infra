# Templates

Jobs and Prow configuration are generated from templates stored in the `templates` directory. The directory contains these items:

- `config.yaml` is the config file that describes files to generate.
- `templates` is the directory to gather all the template files.

### The configuration file structure

The config file has two keys:

- `global` is a map of values available for all templates.
- `templates` is a list of files to generate.

To see the complete structure, see [this](../../development/tools/cmd/rendertemplates/main.go) file.

### Template development

A template receives two objects as input:
- `Values`, which contains all the values specified under `values` in the configuration file.
- `Global` which contains values specified under `global` in the configuration file.

All the functions from [`sprig`](https://github.com/Masterminds/sprig) library are available in the templates. It is the same library that is used by helm, so if you know helm, you are already familiar with them. Also, a few additional functions are available:
- `releaseMatches <release> <since> <until>` returns a boolean value indicating whether `release` fits in the range. Use `nil` to remove one of the bounds. For example: `releaseMatches {{ $rel }} '1.2' '1.5'` checks if the release `$rel` is younger than `1.2` and older than `1.5`.
- `matchingReleases <all-releases> <since> <until>` returns a list of releases filtered to only those that fit in the range.

### Add a new component

To add a new component, find a `templates` entry for `templates/component.yaml`. Then, add a new entry with your component to the `render` list.  
This example defines a component in the Kyma repository using the `go1.12` buildpack:
```yaml
  - from: templates/component.yaml
    render:
      - to: ../prow/jobs/kyma/components/new-component/new-component.yaml
        values:
          <<: *go_kyma_component_1_12
          path: components/new-component
    ...
```

If the buildpack you want to use is not there yet you have to add it. When you add a new buildpack follow the example of already defined ones.

When writing tests for a new component, use the `tester.GetKymaReleasesSince(<next release>)` function to create release jobs tests. This automatically checks whether new release jobs were created in the release process.

### Change component job configuration

To change component job configuration, follow these steps:
1. In the `config.yaml` file, change the name of the file where the jobs are generated. For example, add the suffix `deprecated`. Change the path to this file in tests accordingly.
2. Add `until: <last release>` to this configuration.
3. Create a new entry with the new configuration. Set the `to` field to point to the file responsible for storing jobs.
4. Add `since: <next release>` to the new entry.

Example: buildpack for the API Controller has changed from `go1.11` to `go.12` in release 1.5. This is the component configuration before the buildpack change:
```yaml
      - to: ../prow/jobs/kyma/components/api-controller/api-controller.yaml
        values:
          <<: *go_kyma_component_1_11
          path: components/api-controller
```

This is what the configuration created after the buildpack change looks like:
```yaml
      - to: ../prow/jobs/kyma/components/api-controller/api-controller.yaml
        values:
          <<: *go_kyma_component_1_12
          path: components/api-controllerf
          since: '1.5'
      - to: ../prow/jobs/kyma/components/api-controller/api-controller-go1.11.yaml
        values:
          <<: *go_kyma_component_1_11
          path: components/api-controller
          until: '1.4'
```

When changing tests, use the function `tester.GetKymaReleasesUntil(<last release>)` in place of `tester.GetAllKymaReleases` to test older releases. Use the function `tester.GetKymaReleasesSince(<next release>)` to create release jobs tests for future releases. This automatically checks whether new release jobs were created when doing release.
