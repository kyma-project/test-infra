# Templates

Jobs and Prow configuration are generated from the templates in the `templates` directory. The directory contains these items:

- `config.yaml` is the config file that describes files to generate.
- `templates` is the directory to gather all the template files.

### The configuration file structure

The config file has two keys:

- `global` is a map of values available for all templates.
- `templates` is a list of files to generate.

To see the complete structure, go [here](../../development/tools/cmd/rendertemplates/main.go).

### Template development

A template receives two objects as input:
- `Values`, which contains all the values specified under `values` in the configuration file
- `Global` which contains values specified under `global` in the configuration file.

All the functions from [`sprig`](https://github.com/Masterminds/sprig) library are available in the templates. It is the same library that is used by helm, so if you know helm, you are already familiar with them. Also, a few additional functions are available:
- `releaseMatches <release> <since> <until>` returns a boolean value indicating whether `release` fits in the range. Use `nil` to remove one of the bounds. For example: `releaseMatches {{ $rel }} '1.2' '1.5'` checks if the release `$rel` is younger than `1.2` and older than `1.5`.
- `matchingReleases <all-releases> <since> <until>` returns a list of releases filtered to only those that fit in the range.

### Add a new component

To add a new component, find a `templates` entry for `templates/component.yaml`. Then, add a new entry with your component to the `render` list.  
This example defines a component in the Kyma repository using go1.12 buildpack:
```yaml
  - from: templates/component.yaml
    render:
      - to: ../prow/jobs/kyma/components/new-component/new-component.yaml
        values:
          <<: *go_kyma_component_1_12
          path: components/new-component
    ...
```

If buildpack you're willing to use is not there yet you have to add it. When you add a new buildpack follow the example of already defined ones.

When writing tests for the new component, use the function `tester.GetKymaReleasesSince(<next release>)` to create release jobs tests. This automatically checks whether new release jobs were created when doing release.

### Change component job configuration

To change component job configuration, follow these steps:
1. In `config.yaml`, change the name of the file where the jobs are generated. For example, add the suffix `deprecated`. Change this path in tests as well.
2. Add `until: <last release>` to this configuration.
3. Create new entry with new configuration. It should generate new jobs to the file used before.
4. Add `since: <next release>` to new entry.

Example: buildpack for api-controller has changed from go1.11 to go.12 in release 1.5. Before the change component configuration looked like this:
```yaml
      - to: ../prow/jobs/kyma/components/api-controller/api-controller.yaml
        values:
          <<: *go_kyma_component_1_11
          path: components/api-controller
```

New configuration will look like follows:
```yaml
      - to: ../prow/jobs/kyma/components/api-controller/api-controller.yaml
        values:
          <<: *go_kyma_component_1_12
          path: components/api-controller
          since: '1.5'
      - to: ../prow/jobs/kyma/components/api-controller/api-controller-go1.11.yaml
        values:
          <<: *go_kyma_component_1_11
          path: components/api-controller
          until: '1.4'
```

When changing tests, use the function `tester.GetKymaReleasesUntil(<last release>)` in place of `tester.GetAllKymaReleases` to test older releases. Use the function `tester.GetKymaReleasesSince(<next release>)` to create release jobs tests for future releases. This automatically checks whether new release jobs were created when doing release.