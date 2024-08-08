# Image Builder

Image Builder is a tool for building OCI-compliant images in an SLC-29-compliant system from a GitHub workflow.
It signs images with a signify service to verify that the image comes from a trusted repository and has not been altered in the meantime.
It pushes images to Google Cloud Artifact Registry.

Key features:
* Automatically provides a default tag, which is computed based on a template provided in `config.yaml`
* Supports adding multiple tags to the image
* Supports pushing the same images to multiple repositories
* Supports caching of built layers to reduce build times
* Supports signing images with signify service
* Supports pushing images to the Google Cloud Artifact Registry

> [!NOTE]
> For more information on Image Builder usage in ProwJobs, see [README_deprecated.md](./README_deprecated.md).

## Quickstart Guide

You can use Image Builder in your GitHub workflow to build an image in an SLC-29-compliant system.
[Here](https://github.com/kyma-project/test-infra/blob/main/.github/workflows/pull-image-builder-test.yml) is an example of a GitHub workflow building an image using Image Builder:

```yaml
name: pull-image-builder-test

on:
   pull_request_target:
      types: [ opened, edited, synchronize, reopened, ready_for_review ]
      paths:
         - ".github/workflows/pull-image-builder-test.yml"
         - ".github/workflows/image-builder.yml"
         - ".github/actions/expose-jwt-action/**"
         - ".github/actions/image-builder/**"

permissions:
   id-token: write # This is required for requesting the JWT token
   contents: read # This is required for actions/checkout

jobs:
   compute-tag:
      runs-on: ubuntu-latest
      outputs:
         tag: ${{ steps.get_tag.outputs.TAG }}
      steps:
         - name: Checkout
           uses: actions/checkout@v4
         - name: Get the latest tag
           id: get_tag
           run: echo ::set-output name=TAG::"v0.0.1-test"
         - name: Echo the tag
           run: echo ${{ steps.get_tag.outputs.TAG }}
   build-image:
      needs: compute-tag
      uses: kyma-project/test-infra/.github/workflows/image-builder.yml@main # Usage: kyma-project/test-infra/.github/workflows/image-builder.yml@main
      with:
         name: test-infra/ginkgo
         dockerfile: prow/images/ginkgo/Dockerfile
         context: .
         env-file: "envs"
         tags: ${{ needs.compute-tag.outputs.tag }}
   test-image:
      runs-on: ubuntu-latest
      needs: build-image
      steps:
         - name: Test image
           run: echo "Testing images ${{ needs.build-image.outputs.images }}"
```

The example workflow consists of three jobs:

1. `compute-tag` - computes the tag for the image. It uses the `get_tag` step output to pass the tag to the `build-image` job.
2. `build-image` - builds the image using the Image Builder reusable workflow.
   It uses the `kyma-project/test-infra/.github/workflows/image-builder.yml@main` reusable workflow.
   It builds the `test-infra/ginkgo` image, using the Dockerfile from the `prow/images/gingko/Dockerfile` path.
   The build context is the current directory which effectively means the repository root.
   It uses the `envs` file to load environment variables.
   The image will be tagged with the tag computed in the `compute-tag` job.
3. `test-image` - tests the image build in the `build-image` job. It uses the `build-image` job output to get the image name.

## Workflow Permissions

The Image Builder reusable workflow requires permissions to access the repository and get the OIDC token from the GitHub identity provider.
You must provide the following permissions to the workflow or the job that uses the reusable workflow:

```yaml
permissions:
   id-token: write # This is required for requesting the OIDC token
   contents: read # This is required for actions/checkout
```

## Supported Events

The Image Builder reusable workflow supports the following GitHub events to trigger a workflow:

* `push` - to build images on push to the specified branch.
* `pull_request_target` - to build images on pull requests.
* `workflow_dispatch` - to manually trigger the workflow.

## Reusable Workflow Reference

The workflow that uses the Image Builder reusable workflow must use the exact reference to the reusable workflow.
The value of the `uses` key must be `kyma-project/test-infra/.github/workflows/image-builder.yml@main`.

```yaml
uses: kyma-project/test-infra/.github/workflows/image-builder.yml@main
```

> [!IMPORTANT]
> Using different references to the reusable workflow will result in an error during the workflow execution.

## Reusable Workflow Inputs

The Image Builder reusable workflow accepts inputs to parametrize the build process.
See the accepted inputs description in the [image-builder reusable workflow](/.github/workflows/image-builder.yml) file.

## Reusable Workflow Outputs

The Image Builder reusable workflow provides outputs to pass the results of the build process.
See the provided outputs description in the [image-builder reusable workflow](/.github/workflows/image-builder.yml) file.

## Supported Image Repositories

Image Builder supports pushing images to the Google Cloud Artifact registries.

- Images built on pull requests are pushed to the dev repository, `europe-docker.pkg.dev/kyma-project/dev`.
- Images built on `push` events are pushed to the production repository, `europe-docker.pkg.dev/kyma-project/prod`.

### Image URI

The URI of the image built by Image Builder is constructed as follows:

```
europe-docker.pkg.dev/kyma-project/<repository>/<image-name>:<tag>
```

Where:

* `<repository>` is the repository where the image is pushed. It can be either `dev` or `prod`, based on the event that triggered the build.
* `<image-name>` is the name of the image provided in the `name` input.
* `<tag>` is the tag of the image provided in the `tags` input or the default tag value.

## Image Signing

Image Builder signs images with a signify service.
By default, Image Builder signs images with the production signify service.
Image signing allows verification that the image comes from a trusted repository and has not been altered in the meantime.

> [!NOTE]
> Image Builder signs images built on the push and workflow_dispatch events only. Images built on the pull_request_target event are not signed.

## Named Tags

Image Builder supports passing the name along with the tag, using both the `-tag` option and the config for the tag template.
You can use `-tag name=value` to pass the name for the tag. 

If the name is not provided, it is evaluated from the value:
 - if the value is a string, it is used as a name directly. For example,`-tag latest` is equal to `-tag latest=latest`
 - if the value is go-template, it will be converted to a valid name. For example, `-tag v{{ .ShortSHA }}-{{ .Date }}` is equal to `-tag vShortSHA-Date=v{{ .ShortSHA }}-{{ .Date }}`

> [!Note]
> When running on the `pull_request_target` event, Image Builder ignores additional tags provided with tags input.
> The image will be tagged only with the default PR-<PR_NUMBER> tag.

## Environment File

The environment file contains environment variables to be loaded in the build.
The file must be in the format of `key=value` pairs, separated by newlines.

## Azure DevOps Backend (ADO)

Image Builder uses the ADO `oci-image-builder` pipeline as a build backend.
That means the images are built, signed and pushed to the Google Cloud Artifact registry in the ADO pipeline.
Image Builder does not build images locally on GitHub runners.