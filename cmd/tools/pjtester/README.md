# Prow Job Tester

## Overview

Prow Job tester is a tool for testing changes to the Prow Jobs' definitions and code running in Prow Jobs. It uses the production Prow instance to run chosen Prow Jobs with changes from pull requests (PRs) without going through multiple cycles of new PRs, reviews, and merges. The whole development can be done within one cycle.

### How It Works

The workhorse for testing Prow Jobs is a tool written in Go called `pjtester`. It's available in the `prow-tools` Docker image.

`pjtester` is executed by the presubmit job. This presubmit job is triggered when something changes under the virtual path `vpath/pjtester.yaml`. A PR with the `pjtester.yaml` file is called a pjtester pull request and presubmit running `pjtester` is called pjtester presubmit or pjtester ProwJob.

`pjtester` expects to find the configuration of Prow Jobs to tests under `vpath/pjtester.yaml`.

By default, `pjtester` disables Prow Job reporting to Slack. To check the test results, consult the [Prow Status](https://status.build.kyma-project.io/) dashboard. You can enable reporting to Slack by setting a parameter **report** in `pjtester.yaml` to `true`.

First `pjtester` loads the ProwJob definition. Details from `pjtester.yaml` and from the Prow Job environment variables are used to construct the specification of the Prow Job to test. Prow distinguishes two types of the ProwJob definition sources. Static ProwJobs are stored in the `test-infra` repository and are loaded from local files. Inrepo ProwJobs are stored in other repositories and are loaded through GitHub API.

If the `pjtester.yaml` file contains the **prConfig** parameter, the provided PR number is used to find and load the test ProwJob definition. It applies to both sources, static and inrepo.

If **prConfig** is not provided, the Prow Job tester checks if the PR with the `pjtester.yaml` file is against the same repository as the ProwJob to test. `pjtester` uses the environment variables created by Prow for the presubmit job, which contains the PR refs and commit hash. If this condition is true, a pjtester pull request is used to find and load the test ProwJob definition.

If none of the conditions are met, `pjtester` uses the `heads/main` refs to load the inrepo test ProwJob definition. If the pjtester pull request is open on a repository other than `test-infra`, the static test ProwJob definition is used. If the pjtester pull request is open on the `test-infra` repository, a pjtester pull request is used to find and load the static test ProwJob definition.

Once the ProwJob definition is found and loaded, `pjtester` generates ProwJob specification. ProwJob name and context reported to GitHub are prefixed with the pjtester prefix. ProwJob refs and extraRefs are set according to the configuration provided in the `pjtester.yaml` file in pjtester ProwJob.

If the `pjtester.yaml` file contains PR numbers in the **prConfigs** parameter, they are used as ProwJob refs and extraRefs.

If **prConfigs** doesn't provide a PR number for refs or some extraRefs, but the pjtester pull request is open on the same repository, it is used in the ProwJob specification as refs or extraRefs.

If some extraRefs are not set in the previous steps, they will be set to values loaded from the source. If ProwJob refs are not set, `pjtester` will set them to match the repository `heads/main` details for postsubmit. Presubmit refs are set to match the latest PR merged to the `main` branch.

For presubmit jobs, Prow requires the PR's head SHA, PR number and author set in the Prow Job refs. In the `pjtester.yaml` file, you can specify a PR number for a repository against which a tested Prow Job is running. If you don't specify it, `pjtester` finds the latest PR merged to the`main` branch and uses its details for the presubmit refs.

Finally, `pjtester` creates the ProwJob Kubernetes object on the production Prow instance. The Prow Job name, for which you triggered the test, is prefixed with `{YOUR_GITHUB_USER}_test_of_ProwJob_`.

Because the `vpath/pjtester.yaml` file is used by `pjtester` only, it must not exist outside the PR. This is why the `pre-vpathgurad` required context is added. It fails whenever the `vpath` directory exists and prevents the PR merge. As soon as the virtual path disappears from the PR, `vpathguard` will allow for the PR merge.


## Usage

Next, you must add the `pjtester.yaml` file to the PR to trigger the `pjtester` execution. `pjtester` is run by `pre-<REPO_NAME>-pjtester` Prow Job.

The `pjtester.yaml` file in the virtual path contains configuration parameters for the `pjtester` tool:

| Parameter name | Required | Description                                                                                             |
|----------------|----------|---------------------------------------------------------------------------------------------------------|
| **pjConfigs**  | Yes      | Map containing tests configuration.                                                                     | Yes |
| **prConfig**   | No       | Map containing PR number with test Prow Job definition. The map can contain only one prNumber.              | Yes |
| **prowJobs**   | Yes      | Map containing the configuration of Prow Jobs to test.                                                  | Yes |
| **pjName**     | Yes      | Name of the Prow Job to test.                                                                           | Yes |
| **report**     | No       | Flag enabling reporting of the Prow Job status to Slack. <br> The default value is `false`.             | No |
| **prConfigs**  | No       | Map containing the numbers of the pull requests to use in test ProwJobs. <br> Used as refs or extraRefs. | No |
| **prNumber**   | No       | PR number to use.                                                                                    | No |

An example of the complete `pjtester.yaml` file:

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
3. Create a PR with your changes and the `pjtester.yaml` file.
4. Watch the result of the `{YOUR_GITHUB_USER}_test_of_ProwJob_{TESTED_PROWJOB'S_NAME}` Prow Job.
5. Push new commits to the PR.
6. Redo steps 4 and 5 until you're happy with the test results.
7. Remove the virtual path directory from the PR.
8. Merge your PR.

### Execution of Any Code Without Review?

This was the main requirement for this tool. However, we did put some security in place. The `pre-main-test-infra-pjtester` Prow Job is running on the `trusted-workload` cluster, where it has everything it needs for successful execution. Every Prow Job to test is always scheduled on the `untrusted-workload` cluster, where no sensitive data exists. As for any other PR from a non-Kyma-organization member, every test must be triggered manually.

To prevent overriding existing GitHub contexts results on open PRs with results of execution of test ProwJobs, pjtester adds its prefix to the context defined in the ProwJob definition. This way the test ProwJobs have always their own context name.

### Things to Remember

If you need new Secrets on workload clusters, ask the Neighbors team to create them for your tests.

`pjtester` cannot wait till your new images are build on the PR. This still requires an extra commit after the image is placed in the registry.

## Development

The source code of `pjtester` and its tests is located in `test-infra/development/tools/pkg/pjtester`.
The main function used in the binary is located in `test-infra/development/tools/cmd/pjtester`.

You can't use `pjtester` to test changes to itself.
