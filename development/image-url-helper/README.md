# Image URL Helper

## Overview

Image URL Helper is a tool that provides the following subcommands:

* The `check` command finds all Helm chart images that don't use the `imageurl` template.
* The `list` command lists all Helm chart images by checking the `values.yaml` files.
* The `missing` command lists all non-existent Helm chart images by checking the `values.yaml` files.
* The `promote` command updates the container registry path and Helm chart images versions in the `values.yaml` files. The subcommand also prints a YAML that can be used by the [Image Syncer](../image-syncer) tool to promote images.
* The `components` command generates a [Component-Descriptor](https://github.com/gardener/component-spec) file for the `kyma` repository from `values.yaml` files.


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

To run the `missing` command, use:
```bash
go run main.go \ 
    --resources-directory={PATH_TO_A_KYMA_RESOURCES_DIRECTORY} \
    missing \
    --exclude-test-images=true \
    --output-format=json
```

To run the `promote` command, use:
```bash
go run main.go \ 
    --resources-directory={PATH_TO_A_KYMA_RESOURCES_DIRECTORY} \
    promote \
    --target-container-registry=eu.gcr.io/example \
    --target-tag=release-1 \
    --dry-run=false
```

To run the `components` command, use:
```bash
go run main.go \ 
    --resources-directory={PATH_TO_A_KYMA_RESOURCES_DIRECTORY} \
    components \
    --component-version={COMPONENT_VERSION} \
    --git-commit={GIT_COMMIT_SHA} \
    --git-branch={GIT_BASE_BRANCH} \
    --output-dir={PATH_TO_COMPONENT_DESCRIPTOR_OUTPUT_DIRECTORY} \
    --repo-context={PATH_TO_DOCKER_REPOSITORY}
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

See the list of flags available for the `list` and `missing` command:

| Name                      | Required | Description                                                                                          |
| :------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **--resources-directory** |   Yes    | Path to the Kyma resources directory.|
| **--output-format**       |    No    | Name of the output format (JSON/YAML).|
| **--exclude-test-images**  |    No    | Exclude test images from the output.|


### Promote command flags

See the list of flags available for the `promote` command:

| Name                      | Required | Description                                                                                          |
| :------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **--resources-directory** |   Yes    | Path to the Kyma resources directory.|
| **--target-container-registry** |    No    | Path of the target container registry.|
| **--target-tag**  |    No    | Name of the target image tags.|
| **--dry-run**  |    No    | The boolean value that controls the dry-run mode. It defaults to `true`.|
| **--sign**  |    No    | The boolean value that sets `sign` value in the output YAML list. It defaults to `true`.|
| **--excludes-list** |    No    | Path to the list of excluded `values.yaml` files.|

### Components command flags

See the list of flags available for the `components` command:

| Name                      | Required | Description                                                                                          |
| :------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **--resources-directory** |   Yes    | Path to the Kyma resources directory.|
| **--component-name** |    No    | Name of the component described in the `Component-Descriptor` file. The default value is `github.com/kyma-project/kyma`.|
| **--component-version** |    Yes    | Version of the component.|
| **--provider** |    No    | Component provider (internal or external). The default value is `internal`.|
| **--git-commit** |    Yes    | Hash of the git commit.|
| **--git-branch** |    Yes    | Name of the base git branch.|
| **--skip-image-hashing** |    No    | Boolean value that controls image tag to hash conversion. The default value is `false`.|
| **--output-dir** |    No    | Path to the `Component-Descriptor` file output directory.|
| **--repo-context** |    No    | Name of the Docker repository to push `Component-Descriptor` to. |
