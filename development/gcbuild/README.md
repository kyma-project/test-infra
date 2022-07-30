# gcbuild

This tool servers as an intelligent wrapper for `gcloud builds submit` to run remote build jobs on Google infrastructure
with setting automated substitutions, that developers can use. It's built generally to reduce the complexity of building our Docker images.

Key features:
* Automatically control over values and presence of required substitutions `$_TAG`, `$_REPOSITORY`, `$_VARIANT`
* Allow concurrent builds of image variants that use the same `cloudbuild.yaml` file
* Provide a way to run static checks for `cloudbuild.yaml` files
* Save `gcloud` command outputs to separate files
* Define set of static variables using required config file
* Support for pushing images to different repository with different tags when running in Prow's presubmit job

## Using config.yaml file

`gcbuild` requires a configuration file to be provided with a set of variables, which are used during execution.
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
`variants.yaml` is the file that contains pre-defined set of same substitutions with different values.
It's very handy when there is a requirement to build an image with different versions of the same binary, eg. for different versions of Kubernetes.

The file has a simple structure:
```yaml
'main':
  KUBECTL_VERSION: "1.24.4"
'1.23':
  KUBECTL_VERSION: "1.23.9"
```

If you want to use this feature, you have to make sure that:
* You have `variants.yaml` file in the **same directory**, as `cloudbuild.yaml` file.
* The substitutions you replace with the variants are defined in `cloudbuild.yaml` file and in Dockerfile
* You have `$_VARIANT` substitution defined and used in your image tag, eg. `image:$_TAG-$_VARIANT`

## Linting configuration files

The tool has ability to lint your configuration files before running any workloads on Google Cloud Build.
When you run the tool, it checks your `cloudbuild.yaml` along with `variants.yaml` file, if you meet all requirements in order for it to work.

Additionally, you can use the tool [gcbuild-lint](./tools/gcbuild-lint) if you want to run only the validation checks. This tool is recommended for CI environments.

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
