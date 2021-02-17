# The GitHub Slack Connector for Kyma


## Overview

This document describes the Connectors for GitHub and Slack to use in the [Kyma](https://github.com/kyma-project/kyma) environment. They allow utilizing applications' functions inside the Kyma ecosystem by communicating with the corresponding APIs. Use them to trigger lambda functions on Events incoming from third-party applications and react to them.

## Prerequisites

* **Kyma**
The Connectors are configured to work inside the Kyma ecosystem, so you must install them locally or on a cluster. See the [Installation guides](https://kyma-project.io/docs/root/kyma#installation-installation) for details.

## Usage

You can [install](#quick-start) an example scenario, which labels issues on GitHub that may be offensive and sends notifications to Slack about it. However, considering the fact that the Connectors provide a way to communicate with external applications, there are many possible use cases. Using the Connectors is as simple as deploying a new lambda function in Kyma. Check the corresponding [serverless documentation](https://kyma-project.io/docs/components/serverless) to find out more.

This diagram shows the interaction of the components in the described scenario:

![Software architecture image]

## Quick start

You can install the Connectors and start using them in just a few steps. Follow the instructions to install the Connectors and run the described scenario.

1. Add Add-Ons configuration to Kyma. Run:

    ``` shell
    cat <<EOF | kubectl apply -f -
    apiVersion: addons.kyma-project.io/v1alpha1
    kind: ClusterAddonsConfiguration
    metadata:
      name: addons-slack-github-connectors
      finalizers:
      - addons.kyma-project.io
    spec:
      repositories:
        - url: github.com/dekiel/github-slack-connectors//addons/index.yaml
        - url: github.com/dekiel/github-slack-connectors//addons/index-scenario.yaml
    EOF
    ```

2. Connect to the Kyma Console (UI). Go to a Namespace of your choice, then to **Catalog** in the **Service Management** section. Add the Slack Connector, the GitHub Connector, and the Azure Service Broker. Follow the instructions available in these Add-Ons.
3. After provisioning, add the GitHub Issue Sentiment Analysis Scenario.

    >**NOTE:** Keep in mind that all resources created in the previous step must be ready before you proceed. Check their status in **Instances** in the **Service Management** section.

4. Create a new issue on the GitHub repository specified during the GitHub Connector installation to check if everything is configured correctly. After you create the issue, its sentiment is checked and if it is negative, you get a notification on Slack, and the issue is tagged with the `Caution/offensive` label.

## Installation

Install the Connectors locally or on a cluster. For installation details, see the corresponding guides:

* [The GitHub Connector installation]
* [The Slack Connector installation]

## Development

1. Fork the repository in GitHub.
2. Clone the fork to your `$GOPATH` workspace. Use this command to create the folder structure and clone the repository under the correct location:

    ``` shell
    git clone git@github.com:{GitHubUsername}/github-slack-connectors.git $GOPATH/src/github.com/dekiel/github-slack-connectors
    ```

    Follow the steps described in the [`git-workflow.md`](https://github.com/kyma-project/community/blob/master/contributing/03-git-workflow.md) document to configure your fork.
3. Install dependencies in the main project directory. For example, for the GitHub Connector run:

    ``` shell
    cd github-connector
    dep ensure -vendor-only
    ```


# ProwJob tester

## Overview

ProwJob tester is a tool for testing changes to ProwJobs and scripts in the `test-infra` repository which are under development. It uses the production Prow instance to run chosen ProwJobs with code from pull requests (PRs) without going through multiple cycles of new PRs, reviews, and merges. The whole development is done within one PR.

### How it works

The workhorse for testing ProwJobs is a tool written in Go called `pjtester`. It's available in the `prow-tools` Docker image.

`pjtester` is executed by the presubmit job `pre-master-test-infra-pjtester`. This presubmit job is triggered when something changes under the `test-infra` repository virtual path `vpath/pjtester.yaml`.
It is configured by the **run-if-changed** option:
```bash
run_if_changed: "^(vpath/pjtester.yaml)"
```

`pjtester` expects to find the configuration of the ProwJobs to tests under `vpath/pjtester.yaml`.

By default, `pjtester` disables all reporting for the ProwJob. That means no Slack messages and no status report on Github. To check the test results, consult the [ProwStatus](https://status.build.kyma-project.io/) dashboard.

Details from `pjtester.yaml` and from the ProwJob environment variables are used to construct the specification of the ProwJob to test. `pjtester` uses the environment variables created by Prow for the presubmit job which identify the pull request and its commit hash in the `test-infra` repository. The generated ProwJob to test then uses the `test-infra` code from the pull request's head, ensuring that the latest code is under test.

For presubmit jobs, Prow requires the pull request's head SHA, pull request number, and pull request author set in the ProwJob refs. In the `pjtester.yaml file`, you must specify a pull request number for a repository against which a tested Prow job is running. If you don't specify one, `pjtester` will find a pull request for the `master` branch (`HEAD`) and use its details for the presubmit refs.

Finally, `pjtester` creates the ProwJob on the production Prow instance. The ProwJob name for which you triggered the test is prefixed with `{YOUR_GITHUB_USER}_test_of_prowjob_`.

Because the file `vpath/pjtester.yaml` is used by `pjtester` only to know the name of the ProwJob to test, it should not exist outside of the PR. This is why the `pre-master-test-infra-vpathgurad` required context was added. Its simple task is to fail whenever the `vpath` directory exists and to prevent the PR merge. As soon as the virtual path disappears from the PR, `vpathguard` will allow for the PR merge.

## Prerequisites

ProwJob tester is a tool running on Prow. You need a working Prow instance to use it.

## Installation

To make `pjtester` work for you, you need to compile it and build an image with its binary included. This is done by the `post-test-infra-prow-tools` ProwJob. It builds and pushes the `prow-tools` image.
Next, you must add a presubmit job to trigger `pjtester` execution. This is done by the `pre-master-test-infra-pjtester` ProwJob.

## Configuration

The `pjtester.yaml` file in the virtual path contains configuration parameters for the `pjtester` tool:

| Parameter name | Required | Description |
|----------------|----------|-------------|
| **pjNames** | Yes | List containing the configuration of ProwJobs to test. | Yes |
| **pjNames.pjName** | Yes | Name of the ProwJob to test. | Yes |
| **pjNames.pjPath** | No | Path to the ProwJob definition. <br> Must be relative from the `kyma-project` repository. <br> If not provided, the default location for the `kyma-project/test-infra` repository is used. | No |
| **pjNames.report** | No | Flag enabling reporting of the ProwJob status. The default value is `false`. | No |
| **prConfigs** | No | Dictionary containing the numbers of the pull request on repositories other than `test-infra`. <br> `pjtester` uses their code to test the ProwJobs. | No |
| **configPath** | No | Location of the Prow configuration. <br> Defaults to the path used in `kyma-project/test-infra`. | No |

An example `pjtester.yaml` file:

```
pjNames:
  - pjName: "presubmit-test-job"
    pjPath: "test-infra/development/tools/pkg/pjtester/test_artifacts/"
    report: true
  - pjName: "orphaned-disks-cleaner"
  - pjName: "post-master-kyma-gke-integration"
prConfigs:
  kyma-project:
    kyma:
      prNumber: 1212
configPath: "test-infra/prow/custom_config.yaml"
```


## Usage

This is the Prow job tester flow:

1. Create your feature branch with changes to scripts and/or ProwJobs.
2. Create the `vpath/pjtester.yaml` file with the configuration of the ProwJob to test.
3. Create a PR with your changes.
4. Watch the result of the `{YOUR_GITHUB_USER}_test_of_prowjob_{TESTED_PROWJOB'S_NAME}` Prow job.
5. Push new commits to the PR.
6. Redo steps 4 and 5 until you're happy with the test results.
7. Remove the virtual path directory from the PR.
8. Merge your PR.

### Execution of any code without review?

This was the main requirement for this tool. However, we did put some security in place. The `pre-master-test-infra-pjtester` ProwJob is running on the `trusted-workload` cluster, where it has everything it needs for succesful execution. Every ProwJob to test will be scheduled on the `untrusted-workload` cluster, where no sensitive data exists. As for any other PR from a non-Kyma-organization member, every test must be triggered manually.

### Things to remember

If you need new Secrets on workload clusters, ask the Neighbors team to create it for your tests.

`pjtester` cannot wait till your new images are build on the PR. This still requires an extra commit after the image is placed in the registry.

## Development

The source code of `pjtester` and its tests is located in `test-infra/development/tools/pkg/pjtester`.
The main function used in the binary is located in `test-infra/development/tools/cmd/pjtester`.

You can't use `pjtester` to test changes to itself.
