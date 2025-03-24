# Image Builder: Maintenance Guide

Image Builder is a tool for building OCI-compliant images using the Azure DevOps (ADO) pipeline backend.

It can run in two modes. One is the default mode, where Image Builder acts as a client and triggers the ADO pipeline.
In this mode, Image Builder supports running as part of a GitHub Actions workflow.
In the other mode, Image Builder runs as part of the `oci-image-builder` pipeline in the ADO backend.

Image Builder is built and pushed as a container image to the Google Artifact Registry repository.

## Configuration

Image Builder is configured using a global configuration YAML file, a set of environment variables, and command line flags.

### Configuration YAML File

`image-builder` requires a configuration YAML file. The file holds the global configuration for the tool and is maintained by the authors.
Use the `--config` flag to provide a path to the config YAML file.

For more information about available properties in the configuration file, refer to the [config.go](config.go) file.

See an example file:

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

The `--env-file` specifies the path to the file with environment variables to be loaded in the build.
All variables and their values are loaded into the environment before the build starts.
The file must be in the format of `KEY=VALUE` pairs, separated by newlines.
Environment variables are mainly used to provide runtime values and configuration set by the CI/CD system.
They provide details about the context in which the tool is running.

Image Builder uses several environment variables, which can be grouped by their use cases as follows:

- **ADO_PAT**: Personal Access Token used to authenticate with the ADO API.
- **REPO_OWNER**: Used to extract the repository owner for the ADO pipeline.
- **REPO_NAME**: Used to extract the repository name for the ADO pipeline.
- **JOB_TYPE**: Determines the type of the job (presubmit or postsubmit).
- **PULL_NUMBER**: Used when the job type is a presubmit job.
- **PULL_BASE_SHA**: Used to fetch the base commit SHA for the image tag.
- **PULL_PULL_SHA**: Used when the job type is a pull request.
- **CI**: Determines if Image Builder is running inside a CI system. If set to "true", the CI system is determined and the git state is
  loaded accordingly.
- **GITHUB_REPOSITORY**: Used to get the repository name when the CI system is GitHub Actions.
- **GITHUB_EVENT_NAME**: Used to determine the job type when the CI system is GitHub Actions.
- **GITHUB_EVENT_PATH**: Used to get the path to the event JSON file when the CI system is GitHub Actions.
- **GITHUB_SHA**: Used to get the commit SHA when the CI system is GitHub Actions.
- **GITHUB_REF**: Used to get the pull request number when the CI system is GitHub Actions.

> [!NOTE]
> The actual usage of these environment variables may vary depending on the specific configuration and usage of Image Builder.

### Command Line Flags

Command line flags are the main way for developers to configure the tool and provide needed values for the build process.
Check the list and descriptions of the available flags in
the [main.go](https://github.com/kyma-project/test-infra/blob/df945b96654d60f82b9738cd98129191c5e753c8/cmd/image-builder/main.go#L668) file.

## Azure DevOps (ADO) Build Backend

Image Builder is, by default, used to call ADO API and trigger the `oci-image-builder` ADO pipeline.
When using the ADO backend, Image Builder is used as a client collecting values from flags and environment variables and calling ADO API.
Image Builder triggers the `oci-image-builder` pipeline. This pipeline is responsible for processing parameters provided in a call and
building, pushing, and signing an image.
Image Builder passes the tag definitions and values provided by the user as a base64-encoded pipeline **Tags** parameter.
Encoding the value allows for passing special characters in the tag values without the need to escape them.

Image Builder is used as part of the `oci-image-builder` pipeline in the ADO backend, too.
It's used to execute steps responsible for generating image tags and signing images using the Signify service.

Apart from building images using ADO, Image Builder also supports the preview mode. In the preview mode,
Image Builder does not trigger the ADO pipeline but generates a YAML file with the pipeline definition.
Using this mode allows for the validation of the pipeline definition syntax before triggering it.
To use the preview mode, add the `--ado-preview-run=true` flag.
To specify a path to the YAML file with the pipeline definition, use the `--ado-preview-run-yaml-path` flag.

## Image Signing

Image Builder supports signing images with the Signify service, ensuring that images come from trusted repositories and have not been altered.

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
If the job is running in CI (Prow), it picks up the current **org/repo** value from the default Prow variables.
If binary is running outside of CI, the `--repo` flag must be used. Otherwise, the configuration is not used.

Image Builder contains a basic implementation of a notary signer. If you want to add a new signer, refer to
the [`sign`](../../pkg/sign) package, and its code.

> [!NOTE]
> Images are only signed when built on **push**, **schedule**, and **workflow_dispatch** events. Pull request and merge queue images are not signed.

### Sign-Only Mode

Image Builder supports sign-only mode. To enable it, use the `--sign-only` flag.
It signs the images provided in the `--images-to-sign` flag.
It supports signing multiple images at once. The flag can be used multiple times.

## Named Tags

Image Builder supports passing the name along with the tag, using both the `-tag` option and the config for the tag template.
You can use `-tag name=value` to pass the name for the tag.

If the name is not provided, it is evaluated from the value:

- If the value is a string, it is used as a name directly. For example,`-tag latest` is equal to `-tag latest=latest`
- If the value is go-template, it is converted to a valid name. For example, `-tag v{{ .ShortSHA }}-{{ .Date }}` is equal
  to `-tag vShortSHA-Date=v{{ .ShortSHA }}-{{ .Date }}`.

### Parse-Tags-Only Mode

You can use Image Builder to generate tags using pars-tags-only mode. To enable it, use the `--parse-tags-only` flag.
It parses the tags provided in the `--tag`, `--tag-base64` flags and in `config.yaml`. The generated tags are written as JSON to
stdout.

Flag `--tag-base64` is used to pass the base64-encoded, comma-separated list of tags.
The flag value is decoded and parsed as a list of named tags.
The separated tags are expected to follow the same format as the `--tag` flag values.
The flag is used to pass the tag values to the `oci-image-builder` pipeline in the ADO backend.
The flag value is base64-encoded to avoid issues with special characters in the tag values and do not need to escape them.

## Access Golang SAP Internal Modules

You can use Image Builder to build images that require access to Go SAP internal modules. To use this feature, set the `--use-go-internal-sap-modules=true` flag.
The feature is only available for the ADO backend. 

This feature utilizes the `go` vendor mechanism by downloading all the dependencies into the `vendor` directory. There is no option to pass custom credentials for fetching the dependencies; the tool handles it internally.

## Jenkins Integration

> [!NOTE]
> This Jenkins integration is intended exclusively for the SRE team. It should not be used by other teams or projects.

The Image Builder binary supports integration with the SRE Jenkins instance.
This integration allows the tool to collect input data directly from the Jenkins pipeline environment.

### Jenkins-Specific Flags

- **Build Report Path Flag**: The `--build-report-path` flag has been added to allow the tool to write a build report generated by the ADO
  pipeline directly to a specified file. This functionality is exclusively available for the SRE Jenkins integration. It is not exposed or
  supported when running in the GitHub Actions workflow.

### Required Environment Variables

The Jenkins integration relies on a defined contract for sourcing input data with specific environment variables:

| Variable          | Description                                                                                                                                                                                                                                              | Required                                   |
|-------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------|
| `CHANGE_ID`       | Detects if the pipeline is executed for a pull request. Must contain the PR number.                                                                                                                                                                      | Required for pull requests                 |
| `GIT_URL`         | Repository URL from which an image is built. Must match the regex defined in [`extractOwnerAndRepoFromGitURL` function](https://github.com/kyma-project/test-infra/blob/df945b96654d60f82b9738cd98129191c5e753c8/cmd/image-builder/config.go#L123). | Always required                            |
| `GIT_COMMIT`      | Commit SHA: PR HEAD SHA for pull requests, branch HEAD SHA for push events.                                                                                                                                                                              | Always required                            |
| `CHANGE_BRANCH`   | Name of the pull request's base branch.                                                                                                                                                                                                                  | Required for pull requests                 |
| `CHANGE_BASE_SHA` | Base branch commit SHA for the pull request.                                                                                                                                                                                                             | Required for pull requests                 |
| `JENKINS_HOME`    | Indicates Jenkins runtime environment. Presence of this variable signals the tool to run in Jenkins mode (value not tested).                                                                                                                             | Always required for Jenkins pipeline usage |
