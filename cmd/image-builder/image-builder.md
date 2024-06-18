# Image Builder

Image Builder is a tool for building OCI-compliant images.
It builds images using Azure DevOps (ADO) pipeline backend.
It also supports signing images with a signify service to ensure that the image comes from a trusted repository and has not been altered in
the meantime.

Key features:

* Automatically provides a default tag, which is computed based on a template provided in `config.yaml`
* Supports adding multiple tags to the image
* Saves command outputs to separate files
* When running in Prow's presubmit job, supports pushing images to different repositories with different tags
* Supports pushing the same images to multiple repositories
* Supports caching of built layers to reduce build times


## Configuration

Image Builder is configured using a global configuration YAML file, set of environment variables, and command line flags.

### Configuration YAML File

`image-builder` requires a configuration YAML file. The file holds the global configuration for the tool and is maintained by the authors.
Use the `--config` flag to provide a path to the config YAML file.

For more information about available properties in the configuration file, refer to the [config.go](config.go) file.

Here's an example file:

```yaml
registry: eu.gcr.io/kyma-project
reproducible: true
log-format: json
cache:
  enabled: true
  cache-repo: eu.gcr.io/sap-kyma-neighbors-dev/cache
  cache-run-layers: true
```

### Environment Variables

Environment variables are mainly used to provide runtime values and configuration set by the CI/CD system.
They provide details about the context in which the tool is running.

Here is the list of environment variables used by Image Builder:

# TODO: add detailed descriptoin needed variables.

- **REPO_NAME**: The name of the repository with source code to build an image from.
- **REPO_OWNER**: The owner of the repository with source code.
- **JOB_TYPE**: The type of job. This can be either `presubmit` or `postsubmit`. `presubmit` represents a pull request (PR) job,
  and `postsubmit` represents a push job.
- **PULL_NUMBER**: The number of the PR.
- **PULL_BASE_SHA**: The base SHA of the PR or push commit SHA.
- **PULL_PULL_SHA**: The PR head SHA of the PR.
- **ADO_PAT**: The Azure DevOps Personal Access Token. It's used in the `buildInADO` function to authenticate with the ADO API.
- **USE_BUILDKIT**: Determines whether to use BuildKit for building the image. A `buildkit-image-builder` image has this variable set
  to `true` by default.
- **CI**: Determines whether the image builder runs in CI mode.

### Command Line Flags

Command line flags are the main way for developers to configure the tool and provide needed values for the build process.
Check the list and description of the available flags in
the [main.go](https://github.com/kyma-project/test-infra/blob/df945b96654d60f82b9738cd98129191c5e753c8/cmd/image-builder/main.go#L668) file.

## Azure DevOps Build Backend (ADO)

The ADO backend uses Image Builder to call ADO API and trigger the `oci-image-builder` pipeline. This backend is SLC-29 compliant. It
supports signing images with a production signify service. Images built with ADO can be pushed into Kyma Google Cloud artifacts registries.
To build images, the ADO backend uses the `kaniko-project/executor` image.
This backend doesn't support the `--platform` and `--variant` flags. Building images for platforms other than amd64 is not supported.
To use this backend, you need to use Image Builder in a ProwJob. See [Quickstart Guide](#quickstart-guide) for an example ProwJob
definition.

When using the ADO backend, Image Builder is used as a client collecting values from flags and environment variables and calling ADO API.
Image Builder triggers the `oci-image-builder` pipeline. This pipeline is responsible for processing parameters provided in a call and
building, pushing, and signing an image.

Apart from calling ADO API to trigger image build, Image Builder also supports preview mode. In preview mode,
Image Builder does not trigger the ADO pipeline but generates a YAML file with the pipeline definition.
Using this mode allows for the validation of the pipeline definition syntax before triggering it. To use preview mode, add
the `--ado-preview-run=true` flag.
To specify a path to the YAML file with the pipeline definition, use the `--ado-preview-run-yaml-path` flag.

## Image Signing

Image Builder supports signing the images with a pre-defined set of signing services.
Image signing allows verification that the image comes from a trusted repository and has not been altered in the meantime.
You can enable every supported signing service on repository and global levels.

See the following example of sign services configuration in the `config.yaml` file:

```yaml
sign-config:
  enabled-signers:
    '*':
      - default-signify
    org/repo:
      - repo-token-notary
  signers:
    - name: default-signify
      type: notary
      config:
        endpoint: https://notary/sign
        timeout: 5m
        retry-timeout: 10s
        secret:
          path: /path/to/secret/file/signify.yaml
          type: signify
    - name: repo-token-notary
      type: notary
      config:
        endpoint: https://repo-notary/sign
        timeout: 5m
        retry-timeout: 10s
        secret:
          path: /path/to/secret/file/token
          type: token
```

All enabled signers under `'*'` are used globally. Additionally, if a repository contains another signer configuration
in the **org/repo** key, Image Builder also uses this service to sign the image.
If the job is running in CI (Prow), it picks up the current **org/repo** value from the default Prow variables. If binary
is running outside of CI, the `--repo` flag must be used. Otherwise, the configuration will not be used.

Currently, Image Builder contains a basic implementation of a notary signer. If you want to add a new signer, refer to
the [`sign`](../../pkg/sign) package, and its code.

### Sign-Only Mode

Image Builder supports sign-only mode. To enable it, use the `--sign-only` flag.
It signs the images provided in the `--images-to-sign` flag.
It supports signing multiple images at once. The flag can be used multiple times.

## Named Tags

Image Builder supports passing the name along with the tag, using both the `-tag` option and the config for the tag template.
You can use `-tag name=value` to pass the name for the tag.

If the name is not provided, it is evaluated from the value:

- if the value is a string, it is used as a name directly. For example,`-tag latest` is equal to `-tag latest=latest`
- if the value is go-template, it will be converted to a valid name. For example, `-tag v{{ .ShortSHA }}-{{ .Date }}` is equal
  to `-tag vShortSHA-Date=v{{ .ShortSHA }}-{{ .Date }}`

### Parse-Tags-Only Mode

You can use Image Builder to generate tags using pars-tags-only mode. To enable it, use the `--parse-tags-only` flag.
It parses the tags provided in the `--tag` flag and in `config.yaml`. The generated tags are written as JSON to
stdout.

## Environment Variables File

The `--env-file` specifies the path to the file with environment variables to be loaded in the build.