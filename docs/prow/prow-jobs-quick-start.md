# Prow Jobs QuickStart

This document provides an overview of how to quickly start working with Prow jobs.
> **NOTE:** Please be aware that there are two kinds of Prow jobs. If you want to create a job for the component, read [Manage component jobs with templates](./manage-component-jobs-with-templates.md).

1. Fork the [`test-infra`](https://github.com/kyma-project/test-infra) repository and feature a new branch.


2. Jobs are generated from templates. To create a template, add the `<PROW JOB NAME>-data.yaml` file in the `templates/data` directory. The file should look like this:

    ```yaml
    templates:
      - from: templates/generic.tmpl
        render:
          - to: ../prow/jobs/test-infra/stability-checker.yaml
        <CONFIGURATION>
    ```
    In the `<CONFIGURATION>` part, you can specify local Config Sets (**localSets**) and a configuration of a single job (**jobConfig**), where you can define, for example, the name of the job.
    If needed, global Config Sets (**globalSets**) can be added to the `templates/config.yaml` file.
    
    > **NOTE:** Your template file and Prow job must have unique names.
    
    - To learn more about **localSets**, **jobConfig** and **globalSets**, read [Render Templates](https://github.com/kyma-project/test-infra/tree/main/development/tools/cmd/rendertemplates). 
    - You can search for more examples of template files in the `templates/data` directory.


3. Render the template with this command:
    ```bash
    make jobs-definitions
    ```
    
    For more details on how rendering templates works, read [Render Tamplates](https://github.com/kyma-project/test-infra/tree/main/development/tools/cmd/rendertemplates).
    
    > **CAUTION:** Do not change the generated file! Otherwise, the PR won't be merged, because the job checking the generated file will fail.

   
4. Each Prow job must execute a command. You can either specify it directly in the Prow job definition file (`templates/data/<NAME-data.yaml>`), or attach a script file to the Prow job definition file. The second alternative provides broader options.
    ```yaml
    localSets:
      jobConfig_default:
        command: "<SCRIPT_PATH>/<SCRIPT_NAME.sh>"
    ```
    Script files (`.sh`) are stored in `prow/scripts` directory.


5. To test PR in the Kyma repository, create a new file `vpath/pjtester.yaml` in the `test-infra` repository and reference the pipeline name (`<PROW JOB NAME>`).
    ```yaml
    pjNames:
      - pjName: <PROW JOB NAME>
      - pjName: ...
    ```
    For more details on how to use `pjtester`, read the [Prow Job tester](https://github.com/kyma-project/test-infra/blob/main/development/tools/cmd/pjtester/README.md) document.
 
     
6. Create a pull request (PR) to the `test-infra` repository.

    > **NOTE:** It is recommended to keep PRs as draft ones until you're satisfied with the results.

   
7. Run the test with a comment on your `test-infra` pull request (PR), for example, using `/test all`.
   - To learn more about interacting with Prow, read [Interact with Prow](./prow-jobs.md#interact-with-prow).
   - Look also on [prow command help](https://prow.k8s.io/command-help) for more commands.