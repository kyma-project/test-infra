# Image Autobumper

Image Autobumper is a tool for automatically updating the version of a Docker image in a GitHub repository. It is designed to automatically detected the latest version of the image and update the version in the repository based on the provided configuration.

Key features:
* Automatically detects the latest version of the image
* Automatically detects places in the file where image urls are defined
* Automatically updates the version of the image in the files

## Quickstart Guide

You can use Image Autobumper in your GitHub workflow to update the version of a Docker image in a GitHub repository. It autoamtically detects yaml files (both `.yaml` and `.yml`) in the repository. You can configure the tool to update the version of the image in other files with `extraFiles` option in the configuration file.
[Here](https://github.com/kyma-project/test-infra/blob/main/.github/workflows/autobump-images.yml) is an example of a GitHub workflow updating the version of a Docker image using Image Autobumper.
[Here](https://github.com/kyma-project/test-infra/blob/main/configs/autobump-config/test-infra-autobump-config.yaml) is an example of a configuration file for Image Autobumper.

## Supported events

Image Autobumper supports all events that trigger a GitHub Workflow, except the `pull_request` event. That limitation is caused by requirement to fetch the `kyma-bot` token, which is used to access the forked repository. The `pull_request` event does not provide the OIDC token required to authenticate against Google Cloud Secret Manager.

## Reusable workflow reference

The workflow that uses Image Autobumper reusable workflow must use the exact reference to the reusable workflow. The value of the `uses` key must be `kyma-project/test-infra/.github/workflows/image-autobumper.yml@main`.

```yaml
uses: kyma-project/test-infra/.github/workflows/image-builder.yml@main
```

> [!IMPORTANT]
> Using different references to the reusable workflow will result in an error during the workflow execution.

## Reusable workflow inputs

The Image Autobumper reusable workflow accepts inputs to parametrize the build process. See the accepted inputs description in the [image-autobumper reusable workflow](/.github/workflows/image-autobumper.yml) file.
