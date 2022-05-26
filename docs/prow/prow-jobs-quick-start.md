# Prow Jobs QuickStart

This document provides an overview on how to quickly start working with Prow jobs.

1. Fork [`test-infra`](https://github.com/kyma-project/test-infra) repository and feature new branch


2. Create a template `<PROW JOB NAME>-data.yaml` file in `templates/data` directory. The file should look like this:

```yaml
templates:
  - from: templates/generic.tmpl
    render:
      - to: ../prow/jobs/test-infra/stability-checker.yaml
    <CONFIGURATION>
```
```yaml
templates:
  - fromTo:
      - from: templates/generic.tmpl
        to: ../prow/jobs/kyma/skr-aws-upgrade-integration-dev.yaml
    render:
      - localSets: 
          <...>
      jobConfigs:
        - repoName: "kyma-project/kyma"
          jobs:
            - jobConfig:
                name: "skr-aws-upgrade-integration-dev"
          <...>
    <CONFIGURATION>
```
In `<CONFIGURATION>` part you can specify local config sets (**localSets**) and configuration of a single job (**jobConfig**), where e.g., name of the job can be defined.
In needed, global config sets (**globalSets**) can be added in `templates/config.yaml` file.

- To learn more about **localSets**, **jobConfig** and **globalSets**, please refer to more [specific documentation](https://github.com/kyma-project/test-infra/tree/main/development/tools/cmd/rendertemplates). 
- You can search for more examples of template files in `templates/data` directory.

> **NOTE:** Your prow job must have a unique name.

3. Render template with one of those commands:
```bash
go run ./development/tools/cmd/rendertemplates/main.go --config ./templates/config.yaml
```
or 
```bash
make jobs-definitions
```

- For more details on how rendering templates works, see [this](https://github.com/kyma-project/test-infra/tree/main/development/tools/cmd/rendertemplates) document.

> **NOTE:** Do not change generated file!


4. **Clone https://github.com/kyma-project/test-infra/blob/main/prow/scripts/skr-integration.sh and change name and make sure that make ci-skr correlates to the make file of the fast-int. tests in kyma
   skr-integration.sh
   

5. To test PR in the Kyma repository create a new file `vpath/pjtester.yaml` in the `test-infra` repository
and reference your pipeline name (`<PROW JOB NAME>`) and PR number (`<PR NUMBER>` of `kyma` not `test-infra` repository!).
```
pjNames:
  - pjName: <PROW JOB NAME>
  - pjName: ...
prConfigs:
  kyma-project:
    kyma:
      prNumber: <PR NUMBER> 
```
> **NOTE:** It is recommended to keep PRs as draft ones until you're satisfied with the results.

- For more details on how to use `pjtester`, see [this](https://github.com/kyma-project/test-infra/blob/main/development/tools/cmd/pjtester/README.md)
  document.

6. Run test with comment on your `test-infra` pull request (PR) 
   e.g., using `/test all` 
   - To learn more about interacting with prow, see [this](./prow-jobs.md#interact-with-prow) document.
   - Look also on [prow command help](https://prow.k8s.io/command-help) for more commands.