# image-url-helper

## Overview

The `image-url-helper` tool has two subcommands.
The `check` command is used to find all image usages in Helm charts that doesn't use imageurl template.
The `list` command is used to list all images used in Helm charts by checking values.yaml files.

## Usage

To run `check` command use:
```bash
go run main.go \ 
    --resources-directory={PATH_TO_A_KYMA_RESOURCES_DIRECTORY} \
    check \
    --skip-comments=true \
    --excludes-list={PATH_TO_AN_EXCLUDES_LIST}
```

To run `list` command use:
```bash
go run main.go \ 
    --resources-directory={PATH_TO_A_KYMA_RESOURCES_DIRECTORY} \
    list \
    --exclude-test-images=true \
    --output-format=json
```
# Exclude images from check command
To exclude certain image lines from being checked, provide a path to exclude file in the check command argument. The exclude file contains list of files and values of images excluded from checking:
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


### Check command Flags

See the list of flags available for the `check` command:

| Name                      | Required | Description                                                                                          |
| :------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **--resources-directory** |   Yes    | Path to the Kyma resources directory.|
| **--skip-comments**       |    No    | Skip commented out lines.|
| **--excludes-list**       |    No    | Path to the list of excluded images.|

### List command Flags

See the list of flags available for the `list` commands:

| Name                      | Required | Description                                                                                          |
| :------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **--resources-directory** |   Yes    | Path to the Kyma resources directory.|
| **--output-format**       |    No    | Name of the output format (json/yaml).|
| **--exclude-test-images**  |    No    | Exclude test images from the output.|
