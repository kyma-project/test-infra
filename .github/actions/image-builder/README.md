# Image Builder Action

This action is supposed to trigger image-builder service. It receives authentication tokens (**oidc** and **azure personal access token**) via inputs as well as image name.

Optionally you can set other inputs to control which image with which context should be built.

# Inputs

## **oidc-token**

JWT token exposed by **expose-jwt** github action, which is JWT token signed by github, containing information about running workflow.

Required: **true**

## **ado-token**

Azure DevOps personal access token (PAT), which is used by this action to authenticate against Azure API.

Required: **true**

## **image-name**

Name of the image that will be build via image builder.

Required: **true**

## **context**

Optional parameter that allows you to set build context for image builder, related to the root of the repository.

Default: **'.'**

## **dockerfile**

Optional parameter that allows you to set path to the Dockerfile for the image to build.

Default: **'Dockerfile'**

## **build-args**

Optional parameter that allows you to set additional build arguments.
Each argument has format `name=value`, multiple arguments are separated by new line.

Default: **''**

## **tags**

Optional parameter that allows you to pass tags with which image will be build. Each tag is placed in the new line.

Default: **''**

## **export-tags**

Optional parameter that enable exporting tags provided via **tags** input and default tags to be exported as build argument.
Each tag will have format **TAG_x**, where `x` is the tag name passed along with the tag.

Default: **false**

## **config**

Optional parameter which sets the config file containing information about connection to the Azure DevOps.

Default: **'./configs/kaniko-build-config.yaml'**

## **env-file**

Optional parameter that allows you to provided path to the file with environment variables to be loaded in the build

Default: **''**

## **dry-run**

Optional parameter that allows you to prevent calling Azure service

Default: **false**

# Outputs

## **adoResult**

Result status of the Azure DevOps execution

## **images**

JSON formatted array containing all images build during process.

# How It Work?

Github action is using **europe-docker.pkg.dev/kyma-project/prod/image-builder** docker image to trigger pipeline in Azure DevOps (ADO). It's passing parameters using REST API provided by ADO and kaniko-build-config from `main` branch of `test-infra` repository.

During the ADO pipeline execution is checking for status to be reported, if execution ended it fetches status and logs.

If execution failed, the github action is also failing. If the execution successed, it sets the output and success.

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
          env-file: '.env'
          config: "./configs/kaniko-build-config.yaml"
```