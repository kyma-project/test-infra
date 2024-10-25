# Image Autobumper

Image Autobumper is a tool for automatically updating the version of a Docker image in a GitHub repository. It is designed to automatically detected the latest version of the image and update the version in the repository based on the provided configuration.

Key features:
* Automatically detects the latest Docker image version  in a GitHub repository
* Automatically detects places in the file where image URLs are defined
* Automatically updates the image version in the repository based on the provided configuration

## Quickstart Guide

Use Image Autobumper in your GitHub workflow to detect YAML files (both `.yaml` and `.yml`) in the repository. With the `extraFiles' option in the configuration file, you can also configure the tool to update the version of the image in other files.
See the following examples:
* [GitHub workflow updating the Docker image version using Image Autobumper](https://github.com/kyma-project/test-infra/blob/main/.github/workflows/autobump-images.yml)
* [Configuration file for Image Autobumper](https://github.com/kyma-project/test-infra/blob/main/configs/autobump-config/test-infra-autobump-config.yaml)

## Supported Events

Image Autobumper supports all events that trigger a GitHub workflow except the **pull_request** event. That limitation is caused by the requirement to fetch the `kyma-bot` token, which is used to access the forked repository. The **pull_request** event does not provide the OIDC token required to authenticate against Google Cloud Secret Manager.

## Reusable Workflow Reference

The workflow that uses Image Autobumper reusable workflow must use the exact reference to the reusable workflow. The value of the **uses** key must be `kyma-project/test-infra/.github/workflows/image-autobumper.yml@main`.

```yaml
uses: kyma-project/test-infra/.github/workflows/image-builder.yml@main
```

> [!WARNING]
> Using different references to the reusable workflow results in an error during the workflow execution.

## Reusable Workflow Inputs

The Image Autobumper reusable workflow accepts inputs to parametrize the build process. See the accepted inputs description in the [`image-autobumper` reusable workflow](/.github/workflows/image-autobumper.yml) file.
