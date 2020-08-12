# Prowjob tester

Prowjob tester is a tool for testing changes to Prowjobs and scripts in the `test-infra` repository which are under development. It uses the production Prow instance to run chosen Prowjobs with code from pull requests (PRs) without going through multiple cycles of new PRs, reviews, and merges. The whole development is done within one PR.

## How it works

A workhorse for testing prowjobs is a tool written in Go called `pjtester`. It's available in the `prow-tools` Docker image.

PJtester is called by the presubmit job `pre-master-test-infra-pjtester`. This presubmit is triggered when something changes under the `test-infra` repository virtual path `vpath/pjtester.yaml`.
`run_if_changed: "^(vpath/pjtester.yaml)"`

PJtester expects to find the file with the name of the ProwJob to test in the location `vpath/pjtester.yaml`. Apart from the mandatory ProwJob name, the file may contain the path to the ProwJobs definitions and the path to the Prow configuration. Both paths are optional and are relative from the `kyma-project` directory. If not provided, default locations for the `kyma-project/test-infra` repository are used.

An example `pjtester.yaml` file:

```
pjNames:
  - pjName: "tested-prowjob"
  - pjPath: "test-infra/prow/custom_jobs.yaml"
configPath: "test-infra/prow/custom_config.yaml"
```

Data from `pjtester.yaml` and from the Prowjob environment variables is used to construct the Prowjob specification to test. PJtester will use the environment variables created by Prow for the presubmit which identify the pull request and its commit hash. The generated Prowjob to test will use the `test-infra` code from the pull request's head, ensuring that the latest code is under test.

Finally, PJtester will create the Prowjob on the production Prow instance. The Prowjob name for which you triggered the test will be prefixed with `test_of_prowjob_`.

Because the file `vpath/pjtester.yaml` is used by PJtester only to know the ProwJob name to test, it should not exist outside of the PR. This is why the `pre-master-test-infra-vpathgurad` required context was added. Its simple task is to fail whenever the `vpath` directory exists and to prevent the PR merge. As soon as the virtual path disappears from the PR, `vpathguard` will allow for the PR merge.

## Execution of any code without review?

This was the main requirement for this tool. However, we did place some security in place. The `pre-master-test-infra-pjtester` Prowjob is running on the `trusted-workload` cluster, where it has everything it needs for succesful execution. Every Prowjob to test will be scheduled on the `untrusted-workload` cluster where no sensitive data exists. As for any other PR from a non-Kyma-org member, every test has to be triggered manually.

## How to use it

This is the PJtester flow:

1. Create your feature branch with changes to scripts and/or Prowjobs.
2. Create the `vpath/pjtester.yaml` file with the name of Prowjob to test.
3. Create a PR with your changes.
4. Watch the result of the `test_of_prowjob_{TESTED_PROWJOB'S_NAME}` Prowjob.
5. Push new commits to the PR.
6. Redo steps 4 and 5 until you're happy with the test results.
7. Remove the vpath directory from the PR.
8. Merge your PR.

## Things to remember

If you need new Secrets on workload clusters, you need to ask the Neighbors team to create it for your tests.

PJtester is not able to wait till your new images are build on a PR. That still requires an extra commit after the image is placed in the registry. We will think on how to solve that in [#2632](https://github.com/kyma-project/test-infra/issues/2632).
