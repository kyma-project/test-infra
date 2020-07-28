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

A data from pjtester.yaml and from prowjob environment variables are used to construct prowjob specification to test. PJtester will use environment variables created by Prow for every presubmit which identify pull request and it's commit hash. Generated prowjob to test will use test-infra code from pull request head, ensuring latest code is under test.