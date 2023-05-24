# Terraform Executor

The terraform-executor Task initializes Terraform state locally and runs actions for the provided terraform config.
The usual actions are `plan` and `apply`. Authentication to Google Cloud is done through the Kubernetes service account with the
workload identity configured. The parameter **additional_terraform_args** allows defining Terraform CLI arguments in pipelines.
When this parameter value is set in a pipeline, you must add the default value `-no-color` to the list.
The task uses the tfcmt tool to run Terraform actions and post the results to GitHub PRs.
The tfcmt tool uses github-comments metadata in GitHub comments, so github-comments can be used later for processing tfcmt comments.

## Compatibility

- **Tekton** v0.36.0 and above

## Install

```shell
kubectl apply -f https://raw.githubusercontent.com/kyma-project/test-infra/main/task/terraform-executor/0.1/terraform-executor.yaml
```

## Workspaces

- **repo**: The workspace stores sources for building an image. Usually, this workspace should be shared with other
  tasks. _(REQUIRED)_

## Parameters

- **terraform_action**: Terraform action to execute on provided config files. _(REQUIRED)_
- **module_path**: Path to the Terraform config files. _(REQUIRED)_
- **additional_terraform_args**: Additional Terraform arguments. Add `-no-color` flag if you override the default value. _(OPTIONAL, default: '[ "-no-color" ]')_
- **additional_tfcmt_args**: Additional tfcmt arguments. Allows defining different set of arguments depending on the use case. Parameter must provide at least -sha and/or -pr arguments with values._(REQUIRED)_
- **github-token-secret**: Name of the secret holding the Kyma bot GitHub token. _(OPTIONAL, default: 'kyma-bot-github-token')_
- **REPO_OWNER**: The GitHub organization that triggers the job. A variable set by Prow.  _(REQUIRED)_
- **REPO_NAME**: The GitHub repository that triggers the job. A variable set by Prow.  _(REQUIRED)_

## Platforms

You can run the Task on the `linux/amd64` platform.

## Usage

See the following samples for usage:

- [`prowjob-building-image.yaml`](../samples/sample_prowjob_pipeline.yaml): The presubmit ProwJob that builds, signs, and pushes an image.

## Contributing

We ❤ contributions.

This task is maintained in the [`Test Infra`](https://github.com/kyma-project/test-infra) repository. Issues, pull requests and other contributions can be made there.

To learn more, read the [CONTRIBUTING][contributing] document.

[contributing]: https://github.com/kyma-project/test-infra/blob/main/CONTRIBUTING.md
