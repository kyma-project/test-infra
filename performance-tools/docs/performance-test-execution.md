# Performance Cluster

## Overview

You can execute the performance cluster script in two modes:
* Production mode (executed periodically)
* Development mode (executed on demand)

Are a set of scripts that deploy on demand a kyma cluster on GCP.

### Production Mode
In the **production mode**,  the script is executed periodically. The `runner.sh` script creates a Kyma cluster that needs to be tested. Once the cluster is created, it runs all the K6 scripts and then deletes the cluster.

### Development mode
In the **development mode**, you can create your own Kyma cluster on demand and then run K6 scripts manually. You can use this mode to develop or debug K6 scripts. One can execute the script (cluster.sh)[performance-tools/performance-cluster/cluster.sh] in the following way

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


## Development workflow

>**NOTE:** If you use Dockerhub to push the Kyma Installer image, log in to Docker before you start the `cluster.sh` script.

1. Set up these environment variables:

- **REPO_OWNER** which is the repository owner. 
- **REPO_NAME** which is the repository name. 

2. Create the development cluster:
  ```bash
  ./cluster.sh --action create --cluster-grade development
  ```
3. Once the cluster is ready, check its status:
  ```bash
    kubectl -n default get installation/kyma-installation -o jsonpath="{'Status: '}{.status.state}{', description: '}{.status.description}"; echo; \
  ```
  Check for: `Status: Installed, description: Kyma installed`
4. Run the K6 test scripts:
  ```bash
  k6 run {path to the .js file}
  ```

