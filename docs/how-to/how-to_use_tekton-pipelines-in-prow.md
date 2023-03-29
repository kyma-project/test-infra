# Use Tekton Pipelines with Prow

Kyma Prow instance supports defining and using Tekton Pipelines as a workload. This gives the developer the ability to use Kubernetes-native implementation of CI/CD pipelines, where every task is a self-contained set of commands.

**DISCLAIMER: This document is not complete documentation of Tekton's functionality. This guide is meant only to let you start using Tekton with Prow. For complete information about Tekton's features, read [Tekton official documentation](https://tekton.dev/docs/).**

Tekton can't trigger pipelines natively without an external hook. Prow acts as such a trigger and executes pipeline definition when it receives a webhook event from GitHub, just like ProwJob execution does.

## Defining a PipelineRun in ProwJob

ProwJob's API defines a [`TektonPipelineRunSpec`](https://github.com/kubernetes/test-infra/blob/24e9f4faa85b8504dc8d30b11534a21547159c1e/prow/apis/prowjobs/v1/types.go#L201) specification which is compatible with `v1beta1` API version of Tekton PipelineRunSpec.

> **NOTE:** API `v1beta1` is a relatively new addition to Prow. Some bugs and missing features are expected! Raise an issue with the `kind/bug` label if some functionality is not working.

To get further understanding, see the defined simple presubmit ProwJob that uses Tekton Pipelines as an agent:
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
            workspaces:
              - name: artifacts
                emptyDir: {}
```

As you can see, there are some differences between usual ProwJob, and this one. Every Tekton-based ProwJob has to define the following values as a requirement:
```yaml
decorate: false # Decoration config only applies to Kubernetes-based ProwJobs
agent: tekton-pipeline # It's a requirement to tell Prow to use Tekton as an agent
cluster: tekton # Name of the cluster, where Tekton is working
```

Then you must define a PipelineRun spec using `pipeline_run_spec`. For the information on which fields you can define in this spec, see the [Godoc of `PipelineRunSpec`](https://pkg.go.dev/github.com/tektoncd/pipeline@v0.44.0/pkg/apis/pipeline/v1beta1#PipelineRunSpec).

According to the documentation, you must define a PipelineRunSpec. Dynamically defined `PipelineSpec`, as supported by Tekton's API, is used here:
```yaml
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

The full ProwJob definition looks like this:
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
          params:
            - name: some-param
              value: "Prow+Tekton"
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
                        echo "Bye, $(params.some-param)!"
```

## Reusability of tasks and pipelines

Kyma's Tekton tasks are under [`task` directory](../../task). You can reuse any of the tasks in your own ProwJob definition.

### Reusing tasks

Task defined below would return a "Hello World from task!" to the results field:
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

Tekton supports reusing tasks in Pipelines using the special field `taskRef`. This field is mutually exclusive with `taskSpec`.

Look at the example below that reuses this task. Firstly, the Pipeline will execute the referenced task. Secondly, it will fetch the result of the previous task and print it to the stdout:
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

With this, you can use any pre-defined tasks as many times as you want without defining steps for each one all over again.

### Reusing pipelines

Following the example above, let's define a simple pipeline that returns "Hello world!" and reuses previously defined Task:
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

Then, create a ProwJob that reuses the Pipeline above. To reuse this pipeline, you must use the `pipelineRef` field, which is mutually exclusive to the `pipelineSpec` field.
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

This uses a pre-defined pipeline that is in the Tekton cluster.

## Get repository source in a Pipeline

When applying any Pipeline through Prow, a list of parameters will be propagated in `PipelineRun`.
This list contains fields commonly found in normal ProwJobs. You can check this list in [official Prow docs](https://docs.prow.k8s.io/docs/jobs/#job-environment-variables).

To get the source code, you can use a general purpose [`git-clone`](https://hub.tekton.dev/tekton/task/git-clone) task, or use a [`clone-refs`](https://github.com/kyma-project/test-infra/blob/main/task/clone-refs/0.1/clone-refs.yaml) task which is tailored to use with Prow.

You can reuse git-clone Task available in our [Tekton instance](https://tekton.build.kyma-project.io/#/namespaces/default/tasks/git-clone?view=overview).
Follow the official [Tekton documentation](https://tekton.dev/docs/how-to-guides/clone-repository/) to use it in your pipelines.

To learn how to use the clone-refs task check its [documentation](https://github.com/kyma-project/test-infra/blob/main/task/clone-refs/0.1/README.md) and sample usage in the [prowjob](https://github.com/kyma-project/test-infra/blob/main/task/clone-refs/0.1/samples/prowjob-cloning-repositories.yaml).

To access parameters, use the `$(params.{PARAM})` directive in your scripts or params, where `{PARAM}` is the name of Prow's standard fields from Prow docs.

## Known bugs

Here's a list of know bugs that are most likely to be fixable in upstream Prow.

* ~~Currently, it's impossible to define a custom list of parameters to the Tekton PipelineRun spec defined in a ProwJob. Prow uses this field to provide information about the Git reference on which the Pipeline has been run.~~
* ~~It's impossible to define params in inline tasks. Prow's validation flow returns an incorrect error `Invalid value: "string": val in body must be of type object: "string"`.~~
* When creating a ProwJob with a Tekton agent and extra_refs defined, it's required to specify a Tekton pipeline resource for each extra_ref. Otherwise, the pipeline will fail validation with the following error: `invalid presubmit job tekton-demo: extra_refs[0] is not used; some resource must reference PROW_EXTRA_GIT_REF_0`. As a workaround just add Tekton pipeline resources for each extra_ref in the pipelineRunSpec. The bug is reported to the kubernetes/test-infra [#29144](https://github.com/kubernetes/test-infra/issues/29144).

Bugs have been identified and are reported to the kubernetes/test-infra repository.
[#28679](https://github.com/kubernetes/test-infra/issues/28679) Currently, a workaround that disables Tekton's PipelineRun validation on ProwJob level has been applied on kyma-prow instance.

## Considerations

Although Tekton Pipelines provide a much more complex solution for building pipelines, it still has some drawbacks:

* As tasks work as separate Pods, this can generate an increased load on the Kubernetes cluster, thus increasing the cost in some scenarios.
* Pipelines can be marginally slower to build and define than simple ProwJobs.

If you want to build your pipeline with Tekton, consider the following:

* Does my workflow have to do some complex stuff?
* Can I cover my requirement with a quick Makefile step?

ProwJobs are best suited for simple tasks, whereas Tekton Pipelines are a great way to implement multitask
scenarios or a complex release, or an E2E pipeline with reusable tasks.
