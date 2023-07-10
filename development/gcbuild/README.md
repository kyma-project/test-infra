# gcbuild

This tool serves as an intelligent wrapper for `gcloud builds submit`. It runs remote build jobs on Google infrastructure with setting automated substitutions, that developers can use. It's built to reduce the complexity of building the Docker images.

Key features:
* Automatically control values and presence of required substitutions `$_TAG`, `$_REPOSITORY`, `$_VARIANT`
* Allow for concurrent builds of image variants that use the same `cloudbuild.yaml` file
* Provide a way to run static checks for the `cloudbuild.yaml` files
* Save the `gcloud` command outputs to separate files
* Define a set of static variables using the required config file
* Support for pushing images to different repositories with different tags when running in Prow's presubmit job

## Using config.yaml file

`gcbuild` requires a configuration file to be provided with a set of variables, which are used during the execution.
A `--config` flag is required, however only `project` field is required for program to work properly.

For more information, refer to the [config.go](./config/config.go) file.

```yaml
project: sample-project
devRegistry: dev.kyma-project.io/dev-registry
stagingBucket: gs://staging-bucket
logsBucket: gs://logs-bucket
tagTemplate: v{{ .Date }}-{{ .ShortSHA }}
```

## Building multiple variants of the same image

The tool also includes the functionality of re-using the same `cloudbuild.yaml` pipeline to concurrently build different variants of the same image.
`variants.yaml` is the file that contains pre-defined set of the same substitutions with different values.
It's very handy when there is a requirement to build an image with different versions of the same binary, for example, for different versions of Kubernetes.

The file has a simple structure:
```yaml
'main':
  KUBECTL_VERSION: "1.24.4"
'1.23':
  KUBECTL_VERSION: "1.23.9"
```

To use this feature, make sure that:
* You have the `variants.yaml` file in the **same directory** as the `cloudbuild.yaml` file.
* The substitutions you replace with the variants are defined in the `cloudbuild.yaml` file and in Dockerfile.
* You have the `$_VARIANT` substitution defined and used in your image tag, for example, `image:$_TAG-$_VARIANT`.

## Linting configuration files

The tool has the ability to lint your configuration files before running any workloads on Google Cloud Build.
When you run the tool, it checks your `cloudbuild.yaml` along with the `variants.yaml` file, to confirm that you meet all the requirements for it to work.

Additionally, you can use the tool [gcbuild-lint](./tools/gcbuild-lint), if you want to run only the validation checks. This tool is recommended for the CI environments.

## Usage

```
Usage of gcbuild:
  -build-dir string
        Path to build directory (default ".")
  -cloudbuild-file string
        Path to cloudbuild.yaml file relative to build-dir (default "cloudbuild.yaml")
  -config string
        Path to application config file
  -log-dir string
        Path to logs directory where GCB logs will be stored (default "/logs/artifacts")
  -silent
        Do not push build logs to stdout
  -variant string
        If variants.yaml file is present, define which variant should be built. If variants.yaml is not present, this flag will be ignored

```
