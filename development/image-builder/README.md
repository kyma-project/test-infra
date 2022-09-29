# image-builder

This tool serves as an intelligent wrapper for `kaniko-project/executor`. It reduces the complexity of building Docker images and removes the need of using Docker in Docker when building images in K8s infrastructure.

Key features:
* automatically provides a default tag, which is computed based on a template provided in `config.yaml`
* ~~allows for concurrent builds of image variants that use the same `Dockerfile`~~ See [Known issues](#known-issues) #1
* supports adding multiple tags to the image
* saves command outputs to separate files
* when running in Prow's presubmit job, supports pushing images to different repositories with different tags 
* supports pushing the same images to multiple repositories
* supports caching of built layers to reduce build times

## Known issues

1. Currently, building different variants of the same image is not working. The issue is tracked in https://github.com/kyma-project/test-infra/issues/5975
2. This tool is still at an early stage of development. It is stable enough as a replacement for `docker build`. However, you can expect bugs and codebase changes.

For any other problems, please raise an [issue](https://github.com/kyma-project/test-infra/issues/new?assignees=&labels=area%2Fci%2C+bug&template=bug-report.md&title=image-builder:%20).

## Use config.yaml file

`image-builder` requires a configuration file to be provided with a set of variables, which are used during the execution.
A `--config` flag is required.

For more information, refer to the [config.go](./config.go) file.

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

## Build multi-architecture images

>**NOTE:** This is an experimental feature that may change in the future.

With the introduction of the experimental BuildKit support, the tool now supports the repeatable flag `--platform`.
You can define multiple platforms you want to build an image for.

You can use all platforms supported by [BuildKit](https://github.com/moby/buildkit/blob/master/docs/multi-platform.md).

If you want to use experimental features, there is a new image with the tag suffix `-buildkit`.

## Build multiple variants of the same image

With `image-builder`, you can reuse the same `Dockerfile` to concurrently build different variants of the same image.
To predefine a set of the same `ARG` substitutions with different values, store them in the `variants.yaml` file .
Use that feature when you need to build an image with different versions of the same binary, for example, for different versions of Kubernetes or Go.

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

## Usage

```
Usage of image-builder:
  -config string
        Path to application config file (default "/config/image-builder-config.yaml")
  -context string
        Path to build directory context (default ".")
  -directory string
        Destination directory where the image is be pushed. This flag will be ignored if running in presubmit job and devRegistry is provided in config.yaml
  -dockerfile string
        Path to Dockerfile file relative to context (default "Dockerfile")
  -log-dir string
        Path to logs directory where GCB logs will be stored (default "/logs/artifacts")
  -name string
        Name of the image to be built
  -silent
        Do not push build logs to stdout
  -tag value
        Additional tag that the image will be tagged
  -platform value
        Only supported with BuildKit. Platform of the image that is built
  -variant string
        If variants.yaml file is present, define which variant should be built. If variants.yaml is not present, this flag will be ignored
```
