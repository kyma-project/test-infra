# Run ProwJobs in KinD or k3d

This document provides brief instructions on how to run ProwJobs in local kind (Kubernetes-in-Docker) or k3d locally.

On a workload cluster, each ProwJob runs as a Pod. Prow toolbox contains the following 2 CLI applications that let you generate
a 1:1 representation of a Pod that would be applied on the workload cluster:

* [mkpj](https://github.com/kubernetes-sigs/prow/tree/main/prow/cmd/mkpj) - generates a ProwJob custom resource (CR) from a Prow config with an interactive shell
* [mkpod](https://github.com/kubernetes-sigs/prow/tree/main/prow/cmd/mkpod) - generates a Pod spec from ProwJob custom resource with an interactive shell

The applications serve as a CLI extension of Prow's internal functions to let you generate, run, and debug any ProwJob.

## Prerequisites

* Prow + ProwJobs config files
* Access to secrets to perform tests, if needed
* Docker daemon
* Kubernetes cluster (k3d or kind)

> Podman users: This guide can also work for Podman, but keep in mind that if the job uses Docker in Docker configuration, it's not possible to start the daemon.

## Procedure

1. Compile mkpj and mkpod.
   ```shell
   go install k8s.io/test-infra/prow/cmd/mkpj@latest
   go install k8s.io/test-infra/prow/cmd/mkpod@latest
   ```
2. Start your local kind or k3d cluster and wait for it to finish.
   ```shell
   kind create cluster
   ``` 
3. Create the required ProwJob service account without any RBACs.
   ```shell
   kubectl create serviceaccount prowjob-default-sa # see prow config.yaml for default name for this user
   ```
4. (optional) Create any of the required secret definitions used in a ProwJob you want to test.
5. Go to the `test-infra` repository and create a ProwJob CR from the config.
   ```shell
   mkpj --config-path prow/config.yaml --job-config-path prow/jobs --job {{ JOB_NAME }} > job.yaml
   ```
   The program will interactively ask for which GitHub ref it should create a job. If it's presubmit, it will ask for a PR number from all PRs.
6. Create a PodSpec from the generated ProwJob CR.
   ```shell
   mkpod ../mkpod --prow-job job.yaml --build-id snowflake --local > pod.yaml
   ```
   The program interactively asks for a local directory where ProwJob should store logs. Then, it asks for a path for each of the secrets used by a ProwJob or uses enter to use the Secret stored in a cluster. 
7. Verify the contents of a Pod and apply the YAML to your cluster.
   ```shell
   kubectl apply -f pod.yaml
   ```
   
After these steps, you should see that a Pod has been created, and you can interactively debug your job.

For more information and additional configurations used in local ProwJob debugging, refer to the help command for each of the above CLI tools.
```shell
mkpj -h
mkpod -h
```
