# Image Builder Action

The image builder action triggers the image-builder service. It receives authentication tokens (**oidc** and **azure personal access token**) from inputs as well as from the image name.

Optionally, you can set other inputs to control which image with which context should be built.

# Inputs
Description of each input is available [here](https://github.com/kyma-project/test-infra/blob/main/.github/actions/image-builder/action.yml#L3-L44).

# Outputs
Description of each output is available [here](https://github.com/kyma-project/test-infra/blob/main/.github/actions/image-builder/action.yml#L46-L52).

# How It Works?

The image builder action uses a **europe-docker.pkg.dev/kyma-project/prod/image-builder** Docker image to trigger a pipeline in Azure DevOps (ADO). It passes parameters using REST API provided by ADO. The definition of parameters passed via REST API is taken from kaniko-build-config located on the `main` branch of `test-infra` repository and github context.

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