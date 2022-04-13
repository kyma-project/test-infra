# Prow Job tester

## Overview

Prow Job tester is a tool for testing changes to Prow Jobs and scripts in the `test-infra` repository which are under development. It uses the production Prow instance to run chosen Prow Jobs with code from pull requests (PRs) without going through multiple cycles of new PRs, reviews, and merges. The whole development is done within one PR.

### How it works

The workhorse for testing Prow Jobs is a tool written in Go called `pjtester`. It's available in the `prow-tools` Docker image.

`pjtester` is executed by the presubmit job `pre-main-test-infra-pjtester`. This presubmit job is triggered when something changes under the `test-infra` repository virtual path `vpath/pjtester.yaml`.
It is configured by the **run-if-changed** option:
```bash
run_if_changed: "^(vpath/pjtester.yaml)"
```

`pjtester` expects to find the configuration of Prow Jobs to tests under `vpath/pjtester.yaml`.

By default, `pjtester` disables all reporting for Prow Job. That means no Slack messages and no status report on Github. To check the test results, consult the [Prow Status](https://status.build.kyma-project.io/) dashboard.

Details from `pjtester.yaml` and from the Prow Job environment variables are used to construct the specification of the Prow Job to test. `pjtester` uses the environment variables created by Prow for the presubmit job which identify the pull request and its commit hash in the `test-infra` repository. The generated Prow Job to test then uses the `test-infra` code from the pull request's head, ensuring that the latest code is under test.

For presubmit jobs, Prow requires the pull request's head SHA, pull request number, and pull request author set in the Prow Job refs. In the `pjtester.yaml file`, you must specify a pull request number for a repository against which a tested Prow Job is running. If you don't specify one, `pjtester` will find a pull request for the `main` branch (`HEAD`) and use its details for the presubmit refs.

Finally, `pjtester` creates the Prow Job on the production Prow instance. The Prow Job name for which you triggered the test is prefixed with `{YOUR_GITHUB_USER}_test_of_prowjob_`.

Because the file `vpath/pjtester.yaml` is used by `pjtester` only to know the name of the Prow Job to test, it should not exist outside of the PR. This is why the `pre-main-test-infra-vpathgurad` required context was added. Its simple task is to fail whenever the `vpath` directory exists and to prevent the PR merge. As soon as the virtual path disappears from the PR, `vpathguard` will allow for the PR merge.

## Prerequisites

Prow Job tester is a tool running on Prow. You need a working Prow instance to use it.

## Installation

To make `pjtester` work for you, you need to compile it and build an image with its binary included. This is done by the `post-test-infra-prow-tools` Prow Job. It builds and pushes the `prow-tools` image.
Next, you must add a presubmit job to trigger `pjtester` execution. This is done by the `pre-main-test-infra-pjtester` Prow Job.

## Configuration

The `pjtester.yaml` file in the virtual path contains configuration parameters for the `pjtester` tool:

| Parameter name | Required | Description |
|----------------|----------|-------------|
| **pjNames** | Yes | List containing the configuration of Prow Jobs to test. | Yes |
| **pjNames.pjName** | Yes | Name of the Prow Job to test. | Yes |
| **pjNames.pjPath** | No | Path to the Prow Job definition. <br> Must be relative from the `kyma-project` repository. <br> If not provided, the default location for the `kyma-project/test-infra` repository is used. | No |
| **pjNames.report** | No | Flag enabling reporting of the Prow Job status. The default value is `false`. | No |
| **prConfigs** | No | Dictionary containing the numbers of the pull request on repositories other than `test-infra`. <br> `pjtester` uses their code to test the Prow Jobs. | No |
| **configPath** | No | Location of the Prow configuration. <br> Defaults to the path used in `kyma-project/test-infra`. | No |

An example `pjtester.yaml` file:

```
pjNames:
  - pjName: "presubmit-test-job"
    pjPath: "test-infra/development/tools/pkg/pjtester/test_artifacts/"
    report: true
  - pjName: "orphaned-disks-cleaner"
  - pjName: "post-main-kyma-gke-integration"
prConfigs:
  kyma-project:
    kyma:
      prNumber: 1212
configPath: "test-infra/prow/custom_config.yaml"
```


## Usage

This is the Prow Job tester flow:

1. Create your feature branch with changes to scripts and/or Prow Jobs.
2. Create the `vpath/pjtester.yaml` file with the configuration of the Prow Job to test.
3. Create a PR with your changes.
4. Watch the result of the `{YOUR_GITHUB_USER}_test_of_prowjob_{TESTED_PROWJOB'S_NAME}` Prow Job.
5. Push new commits to the PR.
6. Redo steps 4 and 5 until you're happy with the test results.
7. Remove the virtual path directory from the PR.
8. Merge your PR.

### Execution of any code without review?

This was the main requirement for this tool. However, we did put some security in place. The `pre-main-test-infra-pjtester` Prow Job is running on the `trusted-workload` cluster, where it has everything it needs for succesful execution. Every Prow Job to test will be scheduled on the `untrusted-workload` cluster, where no sensitive data exists. As for any other PR from a non-Kyma-organization member, every test must be triggered manually.

### Things to remember

If you need new Secrets on workload clusters, ask the Neighbors team to create it for your tests.

`pjtester` cannot wait till your new images are build on the PR. This still requires an extra commit after the image is placed in the registry.

## Development

The source code of `pjtester` and its tests is located in `test-infra/development/tools/pkg/pjtester`.
The main function used in the binary is located in `test-infra/development/tools/cmd/pjtester`.

You can't use `pjtester` to test changes to itself.
