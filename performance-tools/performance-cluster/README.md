# Performance Cluster

## Overview

The performance cluster script can be executed in two modes:
* Production Mode (executed periodically)
* Development Mode (executed on Demand)

Are a set of scripts that deploy on demand a kyma cluster on GCP.

### Production Mode
Here the script is executed periodically. It is a kuberentes job defined at (k6pod.yaml)[]. There is runner.sh script which is creating a kyma cluster that needs to be tested. Once the cluster is created, it runs all the k6 scripts present and then deletes the cluster.


### Developement mode
Here a developer can create his/her own kyma cluster on demand and then run k6 scripts manually. The idea here is if a developer wants to develop or debug k6 scripts then he/she can use this mode. One can execute the script (cluster.sh)[] in the following way

```bash
./cluster.sh --action create --cluster-grade developement
```
## Commands

- `action`: It is a required command which indicates the action to be executed for the scripts. Possible action values are `create` or `delete`
- `cluster-grade`: It is a required command which indicates the cluster grade of the kyma cluster. Possible action values are `production` or `development`


Delete Kyma and remove GKE cluster:

- cluster grade development

```bash
./cluster.sh --action delete --cluster-grade development
```

## Expected environment variables:

- DOCKER_REGESTRY
- DOCKER_PUSH_REPOSITORY - Docker repository hostname. Ex. ""
- DOCKER_PUSH_DIRECTORY - Docker "top-level" directory (with leading "/")
   Ex. "/home/${USER}/go/src/github.com/kyma-project"
- CLOUDSDK_CORE_PROJECT - GCP project for all GCP resources used during execution (Service Account, IP Address, DNS Zone, image registry etc.)
- CLOUDSDK_COMPUTE_REGION - GCP compute region. Ex. "europe-west3"
- CLOUDSDK_COMPUTE_ZONE Ex. "europe-west3-a"
- GOOGLE_APPLICATION_CREDENTIALS - GCP Service Account key file path.
  Ex. "/etc/credentials/sa-gke-kyma-integration/service-account.json"
- INPUT_CLUSTER_NAME - name for the new cluster

### Permissions: 

In order to run this script you need to use a service account with permissions equivalent to the following GCP roles:
- Compute Admin
- Kubernetes Engine Admin
- Kubernetes Engine Cluster Admin
- DNS Administrator
- Service Account User
- Storage Admin
- Compute Network Admin


#### Development workflow
* Create the development cluster using following command:
  ```bash
  ./cluster.sh --action create --cluster-grade developement
  ```
* Once the cluster is ready check it using the command
  ```bash
    kubectl -n default get installation/kyma-installation -o jsonpath="{'Status: '}{.status.state}{', description: '}{.status.description}"; echo; \
  ```
  Check for: `Status: Installed, description: Kyma installed`
* One can run the k6 test scripts using following commands
  ```bash
  k6 run <path to the .js file>
  ```

> **NOTE**: Docker container regestry credentials are needed for executing `docker push`. [Authentication methods](https://cloud.google.com/container-registry/docs/advanced-authentication)

