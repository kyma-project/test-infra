# Image Builder Action

The image-builder action triggers the image-builder service. It receives authentication tokens (**oidc** and **azure personal access token**) from inputs as well as from the image name.

Optionally, you can set other inputs to control which image with which context should be built.

# Inputs
Descriptions of each input is available in the [`action`](https://github.com/kyma-project/test-infra/blob/main/.github/actions/image-builder/action.yml#L3-L44) file.

# Outputs
Descriptions of each output is available in the [`action`](https://github.com/kyma-project/test-infra/blob/main/.github/actions/image-builder/action.yml#L46-L52) file.

# How It Works?

The image-builder action uses a **europe-docker.pkg.dev/kyma-project/prod/image-builder** Docker image to trigger a pipeline in Azure DevOps (ADO). It passes parameters using REST API provided by ADO. The definition of parameters passed through REST API is taken from kaniko-build-config located on the `main` branch of the `test-infra` repository and GitHub context.

During the ADO pipeline execution, the image-builder action checks for the status to be reported. When the execution ends, it fetches the status and logs.

If the execution fails, the image-builder action fails. If the execution succeeds, it sets the output and success status.

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