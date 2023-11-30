# Run ProwJobs in KinD or k3d

This document provides brief instructions how to run ProwJobs in local kind (Kubernetes-in-Docker) or k3d locally.

On a workload cluster each ProwJob is running as a Pod. Prow toolbox contains 2 CLI applications that let you generate
a 1:1 representation of a Pod that would be applied on workload cluster.

* [mkpj](https://github.com/kubernetes/test-infra/tree/master/prow/cmd/mkpj) - generates a ProwJob Custom Resource from Prow config with interactive shell
* [mkpod](https://github.com/kubernetes/test-infra/tree/master/prow/cmd/mkpod) - generates Pod spec from ProwJob Custom Resource with interactive shell

Those applications serve as a CLI extension of Prow's internal functions to let anyone generate, run and debug any ProwJob.

## Requirements

* Prow + ProwJobs config files
* Access to secrets to perform tests, if needed
* Docker daemon
* Kubernetes cluster (k3d or kind)

> Podman users: this guide can also work for Podman, but keep in mind that if job uses Docker in Docker configuration, it's not possible to start daemon

## Instruction

1. Compile mkpj and mkpod
   ```shell
   go install k8s.io/test-infra/prow/cmd/mkpj@latest
   go install k8s.io/test-infra/prow/cmd/mkpod@latest
   ```
2. Start local kind or k3d cluster and wait for it to finish
   ```shell
   kind create cluster
   ``` 
3. Create required ProwJob service account without any RBACs
   ```shell
   kubectl create serviceaccount prowjob-default-sa # see prow config.yaml for default name for this user
   ```
4. (optional) Create any of required secret definitions used in a ProwJob you want to test
5. Go into test-infra repository and create a ProwJob CR from config
   ```shell
   mkpj --config-path prow/config.yaml --job-config-path prow/jobs --job {{ JOB_NAME }} > job.yaml
   ```
   The program will interactively ask for which GitHub ref it should create a job. If it's presubmit, it will ask for a PR number from all PRs.
6. Create a PodSpec from generated ProwJob CR
   ```shell
   mkpod ../mkpod --prow-job job.yaml --build-id snowflake --local > pod.yaml
   ```
   The program will interactively ask for a local directory where ProwJob should store logs, then will ask for path for each of the secrets used by a ProwJob, or use enter to use Secret stored in a cluster. 
7. Verify the contents of a Pod and apply the YAML to your cluster
   ```shell
   kubectl apply -f pod.yaml
   ```
   
After these steps you should see that a pod has been created, and you can debug interactively your job.

For more information and additional configurations used in local ProwJob debugging, refer to help command for each of above CLI tools
```shell
mkpj -h
mkpod -h
```