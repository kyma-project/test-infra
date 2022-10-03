# Prow Job tester

## Overview

Prow Job tester is a tool for testing changes to Prow Jobs definitions and code running in Prow Jobs. It uses the production Prow instance to run chosen Prow Jobs with changes from pull requests (PRs) without going through multiple cycles of new PRs, reviews, and merges. The whole development can be done within one cycle.

### How it works

The workhorse for testing Prow Jobs is a tool written in Go called `pjtester`. It's available in the `prow-tools` Docker image.

`pjtester` is executed by the presubmit job. This presubmit job is triggered when something changes under the virtual path `vpath/pjtester.yaml`. Pull request with `pjtester.yaml` file is called a pjtester pull request and presubmit running `pjtester` is called pjtester presubmit or pjtester prowjob.

`pjtester` expects to find the configuration of Prow Jobs to tests under `vpath/pjtester.yaml`.

By default, `pjtester` disables Prow Job reporting to Slack. To check the test results, consult the [Prow Status](https://status.build.kyma-project.io/) dashboard. You can enable reporting to Slack, setting a parameter **report** in pjtester.yaml to true.

First `pjtester` load prowjob definition. Details from `pjtester.yaml` and from the Prow Job environment variables are used to construct the specification of the Prow Job to test. Prow distinct two types of prowjob definition sources. Static prowjobs are stored in `test-infra` repository and are loaded from local files. Inrepo prowjobs are stored in other repositories and are loaded through GitHub API.

If `pjtester.yaml` file contains **prConfig** parameter, provided pull request number is used to find and load test prowjob definition. It applies for both sources, static and inrepo.

If **prConfig** is not provided, Prow Job tester will check if pull request with pjtester.yaml file is against the same repository as prowjob to test. `pjtester` uses the environment variables created by Prow for the presubmit job which contains pull request refs and commit hash. If this condition is true, a pjtester pull request will be used to find and load test prowjob definition.

If none of above conditions are meet, `pjtester` will use `heads/main` refs to load inrepo test prowjob definition and static test prowjob definition if pjtester pull request is open on repository other than `test-infra`. If pjtester pull request is open on `test-infra` repository, a pjtester pull request is used to find and load static test prowjob definition.

Once prowjob definition is found and loaded, `pjtester` generate prowjob specification. Prowjob name and context reported to GitHub is prefixed with pjtester prefix. Prowjob refs and extraRefs are set according to the configuration provided in `pjtester.yaml` file in pjtester prowjob.

If `pjtester.yaml` file contains pull request numbers in **prConfigs** parameter, they will be used as prowjob refs and extraRefs.

If **prConfigs** doesn't provide pull request number for refs or some extraRefs, but pjtester pull request is open on the same repository, it will be used in prowjob specification as refs or extraRefs.

If some extraRefs will not be set in previous steps, they will be set to values loaded from source. If prowjob refs will not be set, `pjtester` will set it to match repository `heads/main` details for postsubmit. Presubmit refs will be set to match latest pull request merged to `main` branch.

For presubmit jobs, Prow requires the pull request's head SHA, pull request number, and pull request author set in the Prow Job refs. In the `pjtester.yaml file`, you can specify a pull request number for a repository against which a tested Prow Job is running. If you don't specify it, `pjtester` will find latest pull request merged to `main` branch and use its details for the presubmit refs.

Finally, `pjtester` creates the ProwJob object on the production Prow instance k8s cluster. The Prow Job name for which you triggered the test is prefixed with `{YOUR_GITHUB_USER}_test_of_prowjob_`.

Because the file `vpath/pjtester.yaml` is used by `pjtester` only, it should not exist outside the PR. This is why the `pre-vpathgurad` required context is added. Its simple task is to fail whenever the `vpath` directory exists and to prevent the PR merge. As soon as the virtual path disappears from the PR, `vpathguard` will allow for the PR merge.


## Usage

Next, you must add a `pjtester.yaml` file to the pull request to trigger `pjtester` execution. `Pjtester` is running by `pre-<REPO_NAME>-pjtester` Prow Job.

The `pjtester.yaml` file in the virtual path contains configuration parameters for the `pjtester` tool:

| Parameter name | Required | Description                                                                                             |
|----------------|----------|---------------------------------------------------------------------------------------------------------|
| **pjConfigs**  | Yes      | Map containing tests configuration.                                                                     | Yes |
| **prConfig**   | No       | Map containing PR number with test Prow Job definition. Map can contain only one prNumber.              | Yes |
| **prowJobs**   | Yes      | Map containing the configuration of Prow Jobs to test.                                                  | Yes |
| **pjName**     | Yes      | Name of the Prow Job to test.                                                                           | Yes |
| **report**     | No       | Flag enabling reporting of the Prow Job status to slack. <br> The default value is `false`.             | No |
| **prConfigs**  | No       | Map containing the numbers of the pull requests to use in test prowjobs. <br> Used as refs or extraRefs. | No |
| **prNumber**   | No       | Number of PR to use.                                                                                    | No |

An example full `pjtester.yaml` file:

```
pjConfigs:
  prConfig:
    kyma-project:
      kyma:
        prNumber: 1313
  prowJobs:
    kyma-project:
      kyma:
      - pjName: "presubmit-test-job"
        report: true
      - pjName: "orphaned-disks-cleaner"
prConfigs: #
  kyma-project:
    kyma:
      prNumber: 1212
```

This is the Prow Job tester flow:

1. Create your feature branch with changes.
2. Create the `vpath/pjtester.yaml` file with the configuration of the Prow Job to test.
3. Create a PR with your changes and `pjtester.yaml` file.
4. Watch the result of the `{YOUR_GITHUB_USER}_test_of_prowjob_{TESTED_PROWJOB'S_NAME}` Prow Job.
5. Push new commits to the PR.
6. Redo steps 4 and 5 until you're happy with the test results.
7. Remove the virtual path directory from the PR.
8. Merge your PR.

### Execution of any code without review?

This was the main requirement for this tool. However, we did put some security in place. The `pre-main-test-infra-pjtester` Prow Job is running on the `trusted-workload` cluster, where it has everything it needs for successful execution. Every Prow Job to test will be always scheduled on the `untrusted-workload` cluster, where no sensitive data exists. As for any other PR from a non-Kyma-organization member, every test must be triggered manually.

To prevent overriding existing GitHub contexts results on open pull requests with results of execution of test prowjobs, pjtester add its prefix to context defined in prowjob definition. This way test prowjobs have always it's own context name.

### Things to remember

If you need new Secrets on workload clusters, ask the Neighbors team to create it for your tests.

`pjtester` cannot wait till your new images are build on the PR. This still requires an extra commit after the image is placed in the registry.

## Development

The source code of `pjtester` and its tests is located in `test-infra/development/tools/pkg/pjtester`.
The main function used in the binary is located in `test-infra/development/tools/cmd/pjtester`.

You can't use `pjtester` to test changes to itself.
