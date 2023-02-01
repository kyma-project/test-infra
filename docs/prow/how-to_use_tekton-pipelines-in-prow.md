# How to use Tekton Pipelines with Prow

Kyma Prow instance supports defininf and using Tekton Pipelines as a workload. This gives the developer ability to use
Kubernetes-native implementation of CI/CD pipelines, where every task is a self-contained set of commands.

**DISCLAIMER: This document is not a complete documentation of Tekton's functionality. This guide is meant only to get you started with using Tekton with Prow. 
If you are looking for complete information about Tekton's features that you can use, refer to [Tekton official documentation](https://tekton.dev/docs/).**

Tekton can't trigger pipelines natively without external hook. Prow acts as such trigger and executes pipeline definition
when it receives a webhook event from GitHub, just as how ProwJob execution works.

## Defining a PipelineRun in ProwJob

ProwJob's API defines a [`TektonPipelineRunSpec`](https://github.com/kubernetes/test-infra/blob/24e9f4faa85b8504dc8d30b11534a21547159c1e/prow/apis/prowjobs/v1/types.go#L201) specification
which is compatible with `v1beta1` API version of Tekton PipelineRunSpec.

> API `v1beta1` is a fairly new addition to Prow, some bugs and missing features are expected! Please, raise an issue if you find some functionality is not working with `kind/bug` label.

To get further understanding let's define simple presubmit ProwJob that uses Tekton Pipelines as an agent:
```yaml
presubmits:
  kyma-project/kyma:
    - name: pull-ci-tekton-run
      decorate: false
      agent: tekton-pipeline
      cluster: tekton-pipelines
      always_run: true
      tekton_pipeline_run_spec:
        v1beta1:
            pipelineRef:
              name: prototype-pipeline
            workspaces:
              - name: artifacts
                emptyDir: {}
```

As you can see, there are some differences between usual ProwJob, and this one. Every Tekton-based ProwJob has to define the following values as a requirement:
```yaml
decorate: false # Decoration config only applies to Kubernetes-based ProwJobs
agent: tekton-pipeline # It's a requirement to tell Prow to use Tekton as an agent
cluster: tekton-pipelines # Name of the cluster, where Tekton is working
```

Then you have to define a PipelineRun spec using `pipeline_run_spec`. For the information which fields you can define in this spec, see the [godoc of `PipelineRunSpec`](https://pkg.go.dev/github.com/tektoncd/pipeline@v0.37.5/pkg/apis/pipeline/v1alpha1#PipelineRunSpec).

According to the documentation, let's define a PipelineRunSpec. I'll be using dynamically defined `PipelineSpec`, as supported by Tekton's API:
```yaml
tekton_pipeline_run_spec:
  v1beta1:
      params:
        - name: username
          value: "Prow+Tekton!"
      pipelineSpec:
        description: "Hello world!"
        tasks:
          - name: hello
            taskSpec:
              steps:
                - name: echo
                  image: alpine:edge
                  script: |
                    #!/bin/sh
                    echo "Hello World"
          - name: bye
            taskSpec:
              steps:
                - name: echo
                  image: alpine:edge
                  script: |
                    #!/bin/sh
                    echo "Bye, $(params.username)"
```

The full ProwJob definition would look like this:
```yaml
presubmits:
  kyma-project/kyma:
    - name: pull-ci-tekton-run
      decorate: false
      agent: tekton-pipeline
      cluster: tekton-pipelines
      always_run: true
      tekton_pipeline_run_spec:
        v1beta1:
          pipelineSpec:
            description: "Hello world!"
            tasks:
              - name: hello
                taskSpec:
                  steps:
                    - name: echo
                      image: alpine:edge
                      script: |
                        #!/bin/sh
                        echo "Hello World"
              - name: bye
                taskSpec:
                  steps:
                    - name: echo
                      image: alpine:edge
                      script: |
                        #!/bin/sh
                        echo "Bye!"
```

## Re-usability of Tasks and Pipelines

Kyma's Tekton pipelines and tasks are under [`configs/tekton/catalog` directory](../../configs/tekton/catalog). You can re-use any of the tasks and pipelines in your own ProwJob definition.

### Re-using tasks

Let's define a simple task that would return a "Hello World from task!" to the results field:
```yaml
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: echo
spec:
  results:
    - name: result
      description: Task result
  steps:
    - name: print-hello-world
      image: bash:latest
      script: |
        #!/usr/bin/env bash
        echo "Hello World from task!" | tee $(results.result.path)
```

Tekton supports re-using tasks in Pipelines using special field `taskRef`. This field is mutually exclusive with `taskSpec`.

Now, let's define a ProwJob that would re-use this task. Firstly the Pipeline will execute referenced task, secondly it will fetch the result of the previous task and print it to the stdout:
```yaml
presubmits:
  kyma-project/kyma:
    - name: pull-ci-tekton-run
      decorate: false
      agent: tekton-pipeline
      cluster: tekton-pipelines
      always_run: true
      tekton_pipeline_run_spec:
        v1beta1:
          pipelineSpec:
            description: "Hello world from external task!"
            tasks:
              - name: hello
                taskRef: 
                  name: echo
```

With this, you can use any pre-defined tasks as many times as you want without needing to define steps for each one all over again.

### Re-using pipelines

Following the example above, let's define a simple pipeline that returns "Hello world!" and re-uses previously defined Task:
```yaml
apiVersion: tekton.dev/v1beta1
kind: Pipeline
metadata:
  name: echo-pipeline
spec:
  description: "Hello world as a Pipeline"
  tasks:
    - name: hello
      taskRef:
        name: echo
    - name: bye
      taskSpec:
        steps:
          - name: echo
            image: alpine:edge
            script: |
              #!/bin/sh
              echo "Buh-Bye!"
```

Then, create a ProwJob that re-uses the Pipeline above. To re-use this pipeline you have to use a `pipelineRef` field, which is mutually exclusive to `pipelineSpec` field.
This field works the same way as `taskRef`:

```yaml
presubmits:
  kyma-project/kyma:
    - name: pull-ci-tekton-run-pipeline
      decorate: false
      agent: tekton-pipeline
      cluster: tekton-pipelines
      always_run: true
      tekton_pipeline_run_spec:
        v1beta1:
          pipelineRef:
            name: echo-pipeline
```

This will use pre-defined pipeline that is in the Tekton cluster. 

## Getting repository source in a Pipeline

When applying any Pipeline though Prow, a list of parameters will be propagated in the `PipelineRun`.
This list contains fields commonly found in normal ProwJobs. You can check this list in [official Prow docs](https://docs.prow.k8s.io/docs/jobs/#job-environment-variables).

To get the source code, simply use [`git-clone`](https://hub.tekton.dev/tekton/task/git-clone) task, or define a similar task in your Pipeline.

To access those parameters use `$(params.{PARAM})` directive in your scripts or params, where `{PARAM}` is a name of Prow's standard fields from Prow docs.

## Known bugs

Here's a list of know bugs that are most likely to be fixable in upstream Prow.

* Currently, it's impossible to define custom list of parameters to the Tekton PipelineRun spec defined in a ProwJob. Prow uses this field to provide information about Git reference the Pipeline has been run on.
* It's impossible to define params in inline tasks. Prow's validation flow returns an incorrect error `Invalid value: "string": val in body must be of type object: "string"`.

## Considerations

Although Tekton Pipelines provide much more complex solution for building pipelines, it still gives some drawbacks:
* YAMLs get utterly cluttered with complex configuration fields
* Understanding pipelines requires good knowledge of Tekton and it's resources
* As Tasks work as separate pods, this will generate increased load on K8s cluster, thus increases the cost
* Pipelines can be marginally slower to build and define, than simple ProwJobs

If you are thinking about building your pipeline with Tekton, then consider the following:
* Does my workflow has to do some complex stuff?
* Can I cover my requirement with quick Makefile step?

If your pipeline just has to run some simple code tests, static checks, work directly on the code and does not require dependencies to external services,
consider using Kubernetes-agent based ProwJob. 

ProwJobs are suitable best for code-oriented tasks, whereas Tekton's Pipelines are a great way to implement a complex release or E2E pipeline with reusable tasks.