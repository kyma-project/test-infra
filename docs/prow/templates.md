# Templates

Jobs and Prow configuration are generated from templates that are in `templates` directory. There are following items there:

- `config.yaml` - config file that describes files to generate
- `templates` - directory to gather all template files

### Config structure

Config file is expected to have two keys:

- `global` - a map of values available for all templates
- `templates` - a list of files to generate

Complete structure can be found [here](../development/tools/cmd/rendertemplates/main.go).

Other then that `config.yaml` uses YAML techniques for better organization, so normal YAML rules apply.

### Template development

A template will receive two objects as inputs:
- `Values` which contains all `values` specified in its configuration in config file
- `Global` which contains values specified under `global` in config file

All functions from [`sprig`](https://github.com/Masterminds/sprig) library are available in templates. It is the same library that is used by helm, so if you know helm there's nothing new. Also some additional functions are available:
- `releaseMatches <release> <since> <until>` - returns a bool value indicating if `release` fits in the range. You can use `nil` to remove one of the bounds. Example: `releaseMatches {{ $rel }} '1.2' '1.5'` will check if release `$rel` is younger then `1.2` and older then `1.5`.
- `matchingReleases <all-releases> <since> <until>` - returns a filtered list of releasesto only those that fits in the range.

### Add new component

To add new component find a `templates` entry for `templates/component.yaml`. Then add new entry with your component to the `render` list. An example below defines a component in Kyma repository using go1.12 buildpack:
```yaml
  - from: templates/component.yaml
    render:
      - to: ../prow/jobs/kyma/components/new-component/new-component.yaml
        values:
          <<: *go_kyma_component_1_12
          path: components/new-component
    ...
```

If buildpack you're willing to use is not there yet you have to add it. The best would be to follow existing buildpacks.