# Image Builder Action

The image builder action triggers the image-builder service. It receives authentication tokens (**oidc** and **azure personal access token**) from inputs as well as from the image name.

Optionally, you can set other inputs to control which image with which context should be built.

# Inputs

## **oidc-token**

A OIDC token exposed by the **expose-jwt** GitHub action, which is a JWT token signed by GitHub, containing information about running workflow.

Required: **true**

## **ado-token**

Azure DevOps personal access token (PAT), which is used by the image builder action to authenticate against the Azure API.

Required: **true**

## **image-name**

Name of the image that is build with the image builder action.

Required: **true**

## **context**

Optional parameter to set the build context for image builder, related to the root of the repository.

Default: **'.'**

## **dockerfile**

Optional parameter to to set the path to the Dockerfile for the image to build.

Default: **'Dockerfile'**

## **build-args**

Optional parameter to set additional build arguments.
Each argument has the format `name=value`, multiple arguments are separated by new line.

Default: **''**

## **tags**

Optional parameter to pass tags with which the image is built. Each tag is placed in a new line.

Default: **''**

## **export-tags**

Optional parameter that enables exporting tags provided with the **tags** input and default tags to be exported as build argument.
Each tag gets the format **TAG_x**, where `x` is the tag name passed along with the tag.

Default: **false**

## **config**

Optional parameter that sets the config file containing information about connection to the Azure DevOps.

Default: **'./configs/kaniko-build-config.yaml'**

## **env-file**

Optional parameter to provide a path to the file with environment variables that loads in the build.

Default: **''**

## **dry-run**

Optional parameter to prevent calling the Azure service

Default: **false**

# Outputs

## **adoResult**

Result status of the Azure DevOps execution

## **images**

Formatted JSON array containing all built images.

# How It Works?

The image builder action uses a **europe-docker.pkg.dev/kyma-project/prod/image-builder** Docker image to trigger a pipeline in Azure DevOps (ADO). It passes parameters using REST API provided by ADO and kaniko-build-config from the `main` branch of `test-infra` repository.

During the ADO pipeline execution, the image builder action is checks for the status to be reported. When the execution ends, it fetches the status and logs.

If the execution failed, the the image builder action fails. If the execution succeeds, it sets the output and success status.

# Example Usage

```
- uses: ./.github/actions/image-builder
        id: build
        name: Run build in image-builder
        with:
          oidc-token: 'oidc-token'
          ado-token: ${{ secrets.ADO_PAT }}
          context: '.'
          build-args: |
            arg1=value1
            arg2=value2
          tags: |
            latest
            main
            commit=${{ .ShortSHA }}
          export-tags: true
          image-name: 'ginkgo'
          dockerfile: 'Dockerfile'
          env-file: 'envs'
          config: "./configs/kaniko-build-config.yaml"
```