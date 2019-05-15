# Performance Cluster

## Overview

The performance cluster script can be executed in two modes:
* Production Mode (executed periodically)
* Development Mode (executed on Demand)

Are a set of scripts that deploy on demand a kyma cluster on GCP.

### Production Mode
Here the script is executed periodically. There is runner.sh script which is creating a kyma cluster that needs to be tested. Once the cluster is created, it runs all the k6 scripts present and then deletes the cluster.

The K6 scripts are located [here](https://github.com/kyma-project/kyma/tests/perf). For more details of how to write K6 scripts in Kyma refer here.

### Development mode
Here a developer can create his/her own kyma cluster on demand and then run k6 scripts manually. The idea here is if a developer wants to develop or debug k6 scripts then he/she can use this mode. One can execute the script (cluster.sh)[https://github.com/kyma-project/test-infra/performance-tools/performance-cluster/cluster.sh] in the following way

```bash
./cluster.sh --action create --cluster-grade development
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

- DOCKER_REGISTRY. Ex. "docker.io"
- DOCKER_PUSH_REPOSITORY - Docker repository hostname. Ex. "docker.io/anyrepository"
- DOCKER_PUSH_DIRECTORY - Docker "top-level" directory (with leading "/")
   Ex. "/home/${USER}/go/src/github.com/kyma-project"
- CLOUDSDK_CORE_PROJECT - GCP project for all GCP resources used during execution (Service Account, IP Address, DNS Zone, image registry etc.)
- CLOUDSDK_COMPUTE_REGION - GCP compute region. Ex. "europe-west3"
- CLOUDSDK_COMPUTE_ZONE Ex. "europe-west3-a"
- GOOGLE_APPLICATION_CREDENTIALS - GCP Service Account key file path.
  Ex. "/etc/credentials/sa-gke-kyma-integration/service-account.json"
- INPUT_CLUSTER_NAME - name for the new cluster
- DOCKER_IN_DOCKER_ENABLED with value "true" for production and "false" for the development flow.

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

> If you are using dockerhub to push Kyma installer image then please login to docker before starting `cluster.sh`

This environment variables are set up by the developer before run the commands.

- **REPO_OWNER** is the repository owner. 
- **REPO_NAME** is the repository name. 

* Create the development cluster using following command:
  ```bash
  ./cluster.sh --action create --cluster-grade development
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

