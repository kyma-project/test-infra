# kaniko-build

This tool serves as an intelligent wrapper for `kaniko-project/executor`. It's built to reduce the complexity of building the Docker images and remove the need of using Docker in Docker when building images in K8s infrastructure.

Key features:
* Automatically provide a default tag, which is computed based on a template provided in `config.yaml`
* ~~Allow for concurrent builds of image variants that use the same `Dockerfile`~~ See Known issues #1
* Support for adding multiple tags to the image
* Save the `kaniko-project/executor` command outputs to separate files
* Support for pushing images to different repositories with different tags when running in Prow's presubmit job
* Support for pushing the same images to multiple repositories
* Supports caching of built layers to reduce build times

## Known issues

1. Currently, building different variants of the same image is not working. The issue is tracked in https://github.com/kyma-project/test-infra/issues/5975
2. This tool is still in early stage of development, though it is stable enough as replacement for `docker build`. Expect bugs and codebase changes!

For any other problems, please raise an [issue](https://github.com/kyma-project/test-infra/issues/new?assignees=&labels=area%2Fci%2C+bug&template=bug-report.md&title=kaniko-build:%20)!

## Using config.yaml file

`kaniko-build` requires a configuration file to be provided with a set of variables, which are used during the execution.
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

## Building multiple variants of the same image

The tool also includes the functionality of re-using the same `Dockerfile` to concurrently build different variants of the same image.
`variants.yaml` is the file that contains pre-defined set of the same `ARG` substitutions with different values.
It's very handy when there is a requirement to build an image with different versions of the same binary, for example, for different versions of Kubernetes or Go.

The file has a simple structure:
```yaml
'main':
  KUBECTL_VERSION: "1.24.4"
'1.23':
  KUBECTL_VERSION: "1.23.9"
```

To use this feature, make sure that:
* You have the `variants.yaml` file in the **same directory** as the `Dockerfile` file.
* Your `Dockerfile` contains `ARG` directives which are named after keys in `variants.yaml`

## Usage

```
Usage of kaniko-build:
  -config string
        Path to application config file (default "/config/kaniko-build-config.yaml")
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
  -variant string
        If variants.yaml file is present, define which variant should be built. If variants.yaml is not present, this flag will be ignored
```
