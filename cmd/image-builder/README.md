# image-builder

`image-builder` is a tool for building OCI-compliant images.
It's able to build images using different backends, such as Kaniko, Buildkit, and Azure DevOps.
It also supports signing images with a pre-defined set of signing services
to verify that the image comes from a trusted repository and has not been altered in the meantime.
The tool is designed to be used in prowjobs.

Key features:
* automatically provides a default tag, which is computed based on a template provided in `config.yaml`
* supports adding multiple tags to the image
* saves command outputs to separate files
* when running in Prow's presubmit job, supports pushing images to different repositories with different tags 
* supports pushing the same images to multiple repositories
* supports caching of built layers to reduce build times

## Quickstart guide

To build an image in an SLC-29 compliant way, use image-builder with ADO backend in your prowjob for building images.
Following is an example of prowjob for building image with ado backend:

```yaml
    - name: pull-build-buildkit-image-builder
      annotations:
        description: "build buildkit image-builder image"
        owner: "neighbors"
      labels:
      run_if_changed: ^pkg/.*.go|cmd/image-builder/.*.go|^go.mod|cmd/image-builder/images/
      decorate: true
      cluster: untrusted-workload # use trusted-workload for postsubmit prowjobs
      max_concurrency: 10
      spec:
        containers:
          - image: "europe-docker.pkg.dev/kyma-project/prod/image-builder:v20240102-18a8a4b8"
            securityContext:
              privileged: false
              seccompProfile:
                type: RuntimeDefault
              allowPrivilegeEscalation: false
            env:
              - name: "ADO_PAT"
                valueFrom:
                  secretKeyRef:
                    name: "image-builder-ado-token"
                    key: "token"
            command:
              - "/image-builder"
            args:
              - "--name=buildkit-image-builder"
              - "--config=/config/kaniko-build-config.yaml"
              - "--context=."
              - "--dockerfile=cmd/image-builder/images/buildkit/Dockerfile"
              - "--build-in-ado=true"
            resources:
              requests:
                memory: 500Mi
                cpu: 500m
            volumeMounts:
              - name: config
                mountPath: /config
                readOnly: true
        volumes:
          - name: config
            configMap:
              name: kaniko-build-config
```

It will build buildkit-image-builder image using image-builder Azure devops backend.
It will use dockerfile from path `cmd/image-builder/images/buildkit/Dockerfile` and config from
configmap `kaniko-build-config`.
Because it's a presubmit prowjob thus it will not sign image.
Signing image is supported only in postsubmit prowjobs.

## Configuration

Image-builder is configured using a global configuration yaml file, set of environment variables, and command line
flags.

### configuration yaml file

`image-builder` requires a configuration yaml file. The file holds global configuration for the tool and is maintained
by authors.
Use `--config` flag to provide a path to the config yaml file.

For more information about available properties in configuration file, refer to the [config.go](config.go) file.

Example file:
```yaml
registry: eu.gcr.io/kyma-project
reproducible: true
log-format: json
cache:
  enabled: true
  cache-repo: eu.gcr.io/sap-kyma-neighbors-dev/cache
  cache-run-layers: true
```

### Environment variables

Environment variables are mainly used to provide a runtime values and configuration set by CI/CD system.
They provide details about the context in which the tool is running.

Following is the list of environment variables used by image-builder:

1. `REPO_NAME`: The name of the repository with source coude to build image from.
2. `REPO_OWNER`: The owner of the repository with srouce code.
3. `JOB_TYPE`: The type of the job. This can be either `presubmit` or `postsubmit`. Presubmit represents a pull request
   job, and postsubmit represents a push job.
4. `PULL_NUMBER`: The number of the pull request.
5. `PULL_BASE_SHA`: The base SHA of the pull request or push commit SHA.
6. `PULL_PULL_SHA`: The pull request head SHA of the pull request.
7. `ADO_PAT`: The Azure DevOps Personal Access Token. It's used in the `buildInADO` function to authenticate with the
   Azure DevOps API.
8. `USE_BUILDKIT`: Determines whether to use BuildKit for building the image. A buildkit-image-builder image has this
   variable set to `true` by default.
9. `CI`: Determines whether the image builder is running in CI mode.

### Command line flags

Command line flags are a main way for developers to configure the tool and provide needed values for the build process.
Check list and description of availables flags
in [main.go](https://github.com/kyma-project/test-infra/blob/df945b96654d60f82b9738cd98129191c5e753c8/cmd/image-builder/main.go#L668)
file.

## Image signing

Image-builder supports signing the images with a pre-defined set of signing services.
Signing of images allows to verify that image comes from a trusted repository and has not been altered in the meantime.
You can enable every supported signing service on repository and global levels.

See the following example sign services configuration in `config.yaml` file:
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
in the `org/repo` key, image-builder also uses this service to sign the image.
If the job is running in CI (Prow), it picks up the current `org/repo` value from the default Prow variables. If binary
is running outside of CI, `--repo` flag will have to be used. Otherwise, the configuration will not be used.

Currently, image-builder contains a basic implementation of a notary signer. If you want to add a new signer, refer to
the [`sign`](../../pkg/sign) package, and its code.

### Sign only mode

Image-builder supports sign-only mode. To enable it, use `--sign-only` flag.
It will sign the images provided in the `--images-to-sign` flag.
It supports signing multiple images at once. The flag can be used multiple times.

## Named Tags

Image-builder supports passing the name along with the tag both using the `-tag` option or config for the tag template.
You can use `-tag name=value` to pass the name for the tag. 

If the name is not provided, it is evaluated from the value:
 - if the value is a string, it is used as a name directly. For example,`-tag latest` is equal to `-tag latest=latest`
 - if the value is go-template, it will be converted to a valid name. For example, `-tag v{{ .ShortSHA }}-{{ .Date }}` is equal to `-tag vShortSHA-Date=v{{ .ShortSHA }}-{{ .Date }}`

### Parse tags only mode

You can use image-builder to generate tags using pars tags only mode. To enable it, use `--parse-tags-only` flag.
It will parse the tags provided in the `--tag` flag and in `config.yaml`. Generated tags will be written as json to
stdout.

## Build backend

Image-builder supports three build backends:

- Kaniko
- Buildkit
- Azure devops pipelines

Kaniko and Buildkit build image locally while Azure devops pipelines backend call ADO API.
To use kaniko backend, use image-builder image.
To use buildkit backend, use buildkit-image-builder image.
Azure devops backend is supported by both images. To use it, you need to provide `--build-in-ado=true` flag.
Buildkit and Kaniko backends are deprecated and will be removed in the future.
The Preferred way to build images is to use ado backend, because it's the only SLC-29 compliant backend.

### Azure devops backend

Azure devops backend uses image-builder to call ADO API and trigger oci-image-builder pipeline. This backend is SLC-29
compliant. It supports signing images with production signify service. Images build with ADO can be pushed into kyma GCP
artifacts registers. To build images ADO backend uses `kaniko-project/executor` image. This backend doesn't
support `--env-file`, `--platform` and `--variant` flags. Building images for platforms other than amd64 is not
supported. To use this backend, you need to use image-builder in a prowjob. See [Quickstart guide](#quickstart-guide)
for
example prowjob definition.

When using ADO backend, an image-builder is used as client collecting values from flags and environment variables and
calling ADO API. Image-builder triggers oci-image-builder pipeline. This pipeline is responsible for processing
parameters provided in a call and building, pushing and signing image.

Apart from calling ADO API to trigger image build, image-builder also supports preview mode. In preview mode,
image-builder will not trigger ADO pipeline but will generate yaml file with pipeline definition. Using this mode allows
to validate syntax of pipeline definition before triggering it. To use preview mode use `--ado-preview-run=true` flag.
To specify a path to yaml file with pipeline definition use `--ado-preview-run-yaml-path` flag.

## Deprecated features

### Build multi-architecture images

> **NOTE:** This is an experimental feature that may change in the future.

With the introduction of the experimental BuildKit support, the tool now supports the repeatable flag `--platform`.
You can define multiple platforms you want to build an image for.

You can use all platforms supported by [BuildKit](https://github.com/moby/buildkit/blob/master/docs/multi-platform.md).

If you want to use experimental features, there is a new image with the tag suffix `-buildkit`.

### Build multiple variants of the same image

With `image-builder`, you can reuse the same `Dockerfile` to concurrently build different variants of the same image.
To predefine a set of the same `ARG` substitutions with different values, store them in the `variants.yaml` file .
Use that feature when you need to build an image with different versions of the same binary, for example, for different
versions of Kubernetes or Go.

The file has a simple structure:

```yaml
'main':
  KUBECTL_VERSION: "1.24.4"
'1.23':
  KUBECTL_VERSION: "1.23.9"
```

To use this feature, make sure that:

* you have the `variants.yaml` file in the **same directory** as the `Dockerfile`
* your `Dockerfile` contains `ARG` directives which are named after keys in `variants.yaml`

### Environment variables file

`-env-file` specify path to the file with environment variables to be loaded in build. This flag is deprecated.
Use `--build-arg` instead.
