# Prowjob tester

It's a tool for testing changes in prowjobs and scripts in test-infra repository which are under development. It uses production Prow instance to run choosen prowjobs with code from pull requests without going through multiple cycles of new PRs, reviews and merges. All development is done within one PR.

## How it works

A workhorse for testing prowjobs is a tool written in go called *pjtester*. It's available in prow-tools docker image.

PJtester is called by presubmit job *pre-master-test-infra-pjtester*. This presubmit is triggered when anything change under the test-infra repository virtual path _vpath/pjtester.yaml_.
`run_if_changed: "^(vpath/pjtester.yaml)"`

PJtester expect to find a file with name of prowjob to test in location _vpath/pjtester.yaml_. Apart from mandatory prowjob name, a file may contain path to prowjobs definitions and path to prow config. Both paths are optional and are relative from kyma-project directory. If not provided default locations for kyma-project/test-infra repository are used.

Example pjtester.yaml file.
```buildoutcfg
pjName: "tested-prowjob"
pjPath: "test-infra/prow/custom_jobs.yaml"
configPath: "test-infra/prow/custom_config.yaml"
```

A data from pjtester.yaml and from prowjob environment variables are used to construct prowjob specification to test. PJtester will use environment variables created by Prow for presubmit which identify pull request and it's commit hash. Generated prowjob to test will use test-infra code from pull request head, ensuring latest code is under test.

Finally pjtester will create prowjob on production Prow instance. A prowjob name for wich you triggered test will be prefixed with _test_of_prowjob__.

Because file _vpath/pjtester.yaml_ is used only by pjtester to know prowjob name to test, it should not exist outside PR. This is the reason pre-master-test-infra-vpathgurad required context was added. Its simple task is to fail whenever _vpath_ directory exist and prevent PR merge. As soon virtual path will disappear from PR, vpathguard will allow PR merge.

## Exec of any code without review?

This was the main requirement for this tool. However, we did place some security in place. Prowjob pre-master-test-infra-pjtester is running on **trusted-workload** cluster where it has everything needed for success execution. Every prowjob to test will be scheduled on **untrusted-workload** cluster where no sensitive data exist. As for any other PR from non Kyma org member, every test have to be triggered manually.

## How to use it

This is a pjtester flow.
1. Create your feature branch with changes in scripts and/or prowjobs.
2. Create vpath/pjtester.yaml file with name of prowjob to test.
3. Create PR with you changes.
4. Watch result of prowjob _test_of_prowjob_<tested prowjob name>_.
5. Push new commits to PR.
6. Redo points 4 and 5 till you're happy with tests result.
7. Remove vpath directory from PR.
8. Merge your PR.

## Things to remember

If you need new secrets on workload clusters, you need to ask neighbors team to create it for your tests.

PJtester is not able to wait till your new images will be build on a PR. That still requires extra commit after image is placed in registry. We will think how to solve that in #2632