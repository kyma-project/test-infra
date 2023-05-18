# Terraform Executor

The terraform-executor task initialise locally terraform state and run actions for provided terraform config.
Usually, actions are plan and apply. Authentication to the Google Cloud is done through k8s service account with
workload identity configured. A parameter additional_terraform_args allows defining terraform cli arguments in pipelines.
When this parameter value is set in pipeline, a default value -no-color must be added to the list.
The task uses tfcmt tool to run terraform actions and post results to GitHubPR.
tfcmt tool use a github-comments metadata in GitHub comments, so github-comments can be used later for processing tfcmt comments.

## Compatibility

- **Tekton** v0.36.0 and above

## Install

```shell
kubectl apply -f https://raw.githubusercontent.com/kyma-project/test-infra/main/task/terraform-executor/0.1/terraform-executor.yaml
```

## Workspaces

- **repo**: The workspace stores sources for building an image. Usually, this should be a workspace shared with other
  tasks. _(REQUIRED)_

## Parameters

- **terraform_action**: Terraform action to execute on provided config files. _(REQUIRED)_
- **module_path**: Path to the terraform config files. _(REQUIRED)_
- **additional_terraform_args**: Additional terraform arguments. Add -no-color flag if you override default value. _(OPTIONAL, default: '[ "-no-color" ]')_
- **github-token-secret**: Name of the secret holding the kyma bot github token. _(OPTIONAL, default: 'kyma-bot-github-token')_
- **PULL_NUMBER**: Pull request number. A variable set by Prow. _(REQUIRED)_
- **SHA**: Commit hash with terraform config files.  _(REQUIRED)_
- **REPO_OWNER**: The GitHub organization that triggers the job. A variable set by Prow.  _(REQUIRED)_
- **REPO_NAME**: The GitHub repository that triggers the job. A variable set by Prow.  _(REQUIRED)_

## Platforms

You can run the Task on `linux/amd64` platform.

## Usage

See the following samples for usage:

- [`prowjob-building-image.yaml`](samples/sample_prowjob_pipeline.yaml): A presubmit ProwJob that builds,signs and pushes an image.

## Contributing

We ❤ contributions.

This task is maintained in the [Test Infra](https://github.com/kyma-project/test-infra) repository. Issues, pull requests and other contributions can be made there.

To learn more, read the [CONTRIBUTING][contributing] document.

[contributing]: https://github.com/kyma-project/test-infra/blob/main/CONTRIBUTING.md
