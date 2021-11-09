# Image URL Helper

## Overview

Image URL Helper is a tool that provides two subcommands:

* The `check` command finds all Helm chart images that don't use the `imageurl` template.
* The `list` command lists all Helm chart images by checking the `values.yaml` files.

## Usage

To run the `check` command, use:
```bash
go run main.go \ 
    --resources-directory={PATH_TO_A_KYMA_RESOURCES_DIRECTORY} \
    check \
    --skip-comments=true \
    --excludes-list={PATH_TO_AN_EXCLUDES_LIST}
```

To run the `list` command, use:
```bash
go run main.go \ 
    --resources-directory={PATH_TO_A_KYMA_RESOURCES_DIRECTORY} \
    list \
    --exclude-test-images=true \
    --output-format=json
```
### Exclude images from the check command
To exclude image lines from being checked, create a YAML file that contains a list of files and values of images that you want to exclude from the check. Then, provide a path to this file in the `check` command argument. See the example of such a YAML file:

```yaml
excludes:
  - filename: "monitoring/charts/grafana/values.yaml"
    images:
      - "bats/bats"
  - filename: "monitoring/charts/grafana/templates/image-renderer-deployment.yaml"
    images:
     - "{{ .Values.imageRenderer.image.repository }}:{{ .Values.imageRenderer.image.tag }}@sha256:{{ .Values.imageRenderer.image.sha }}"
     - "{{ .Values.imageRenderer.image.repository }}:{{ .Values.imageRenderer.image.tag }}"
```


### Check command flags

See the list of flags available for the `check` command:

| Name                      | Required | Description                                                                                          |
| :------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **--resources-directory** |   Yes    | Path to the Kyma resources directory.|
| **--skip-comments**       |    No    | Skip commented out lines.|
| **--excludes-list**       |    No    | Path to the list of excluded images.|

### List command flags

See the list of flags available for the `list` command:

| Name                      | Required | Description                                                                                          |
| :------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **--resources-directory** |   Yes    | Path to the Kyma resources directory.|
| **--output-format**       |    No    | Name of the output format (JSON/YAML).|
| **--exclude-test-images**  |    No    | Exclude test images from the output.|
