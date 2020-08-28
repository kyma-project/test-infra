# ProwJob tester

## Overview

ProwJob tester is a tool for testing changes to ProwJobs and scripts in the `test-infra` repository which are under development. It uses the production Prow instance to run chosen ProwJobs with code from pull requests (PRs) without going through multiple cycles of new PRs, reviews, and merges. The whole development is done within one PR.

### How it works

A workhorse for testing ProwJobs is a tool written in Go called `pjtester`. It's available in the `prow-tools` Docker image.

Pjtester is called by the presubmit job `pre-master-test-infra-pjtester`. This presubmit is triggered when something changes under the `test-infra` repository virtual path `vpath/pjtester.yaml`.

`run_if_changed: "^(vpath/pjtester.yaml)"`

Pjtester expects to find the file with the configuration of ProwJobs tests in the location `vpath/pjtester.yaml`.

A list `pjNames` contains configuration of ProwJobs to test. Each element of a list must contain `pjName` key with name of ProwJob to test. It may contain `pjPath` key with path to the ProwJob definition and `report: true` which enables reporting of Prow job status. Only `pjName` key is mandatory. The path to the ProwJob configuration should be relative from the `kyma-project` directory. If not provided, default location for the `kyma-project/test-infra` repository is used.

A dictionary `prConfigs` contains numbers of pull request on repositories other than test-infra. Pjtester will use code from these PRs to test ProwJobs. It's optional.

Item `configPath` holds location to the configuration of Prow. It defaults to the path used in `kyma-project/test-infra` and is optional.

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

By default, pjtester will disable all reporting for a Prow job. That means no slack messages and no status report on Github will be provided. To check test result please consult [ProwStatus](https://status.build.kyma-project.io/) dashboard.

Details from `pjtester.yaml` and from the Prow job environment variables are used to construct the ProwJob specification to test. Pjtester will use the environment variables created by Prow for the presubmit which identify the pull request and its commit hash on `test-infra` repository. The generated ProwJob to test will use the `test-infra` code from the pull request's head, ensuring that the latest code is under test.

Finally, pjtester will create the ProwJob on the production Prow instance. The Prow job name for which you triggered the test will be prefixed with `{YOUR_GITHUB_USER}_test_of_prowjob_`.

Because the file `vpath/pjtester.yaml` is used by pjtester only to know the ProwJob name to test, it should not exist outside of the PR. This is why the `pre-master-test-infra-vpathgurad` required context was added. Its simple task is to fail whenever the `vpath` directory exists and to prevent the PR merge. As soon as the virtual path disappears from the PR, `vpathguard` will allow for the PR merge.

## Prerequisites

ProwJob tester is a tool running on Prow. You need working Prow instance to use it.

## Installation

To make pjtester work for you, you need to compile it and build image with its binary included. This is done by ProwJob `post-test-infra-prow-tools`. It will build and push `prow-tools` image. Next you need to add presubmit job to trigger pjtester execution. This is done by `pre-master-test-infra-pjtester` ProwJob.

## Usage

This is the Prow job tester flow:

1. Create your feature branch with changes to scripts and/or ProwJobs.
2. Create the `vpath/pjtester.yaml` file with the config of ProwJob to test.
3. Create a PR with your changes.
4. Watch the result of the `{YOUR_GITHUB_USER}_test_of_prowjob_{TESTED_PROWJOB'S_NAME}` Prow job.
5. Push new commits to the PR.
6. Redo steps 4 and 5 until you're happy with the test results.
7. Remove the vpath directory from the PR.
8. Merge your PR.

### Execution of any code without review?

This was the main requirement for this tool. However, we did put some security in place. The `pre-master-test-infra-pjtester` ProwJob is running on the `trusted-workload` cluster, where it has everything it needs for succesful execution. Every ProwJob to test will be scheduled on the `untrusted-workload` cluster where no sensitive data exists. As for any other PR from a non-Kyma-org member, every test has to be triggered manually.

### Things to remember

If you need new Secrets on workload clusters, ask the Neighbors team to create it for your tests.

Pjtester is not able to wait till your new images are build on a PR. That still requires an extra commit after the image is placed in the registry.

## Development

Source code of pjtester and its tests are located in `test-infra/development/tools/pkg/pjtester`.
Main function used in binary is located in `test-infra/development/tools/cmd/pjtester`.

You can't use pjtester to test changes to itself.