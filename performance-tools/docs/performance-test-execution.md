# Performance cluster

The [`cluster.sh`](../performance-cluster/cluster.sh) script is the performance cluster script that you can run in two modes:
* Production mode (executed periodically)
* Development mode (executed on demand)

In the **production mode**, `runner.sh` periodically executes the `cluster.sh` script. This script creates a Kyma cluster that needs to be tested. Once the cluster is created, it runs all the [K6 scripts](https://github.com/kyma-project/kyma/tree/master/tests/perf) and then deletes the cluster.

A Kubernetes job defined in [`performance-tests-k6-runner-cronjob.yaml`](../performance-cluster/job/performance-tests-k6-runner-cronjob.yaml) triggers the execution of [`runner.sh`](../performance-cluster/runner.sh).

>**NOTE:** For more details of how to write K6 scripts in Kyma, see [this document](https://github.com/kyma-project/kyma/tree/master/tests/perf/README.md).

In the **development mode**, you can create your own Kyma cluster on demand and then run K6 scripts manually. You can use this mode to develop or debug K6 scripts.

In order to run the script, you need a service account with permissions equivalent to the following GCP roles:
- Compute Admin
- Kubernetes Engine Admin
- Kubernetes Engine Cluster Admin
- DNS Administrator
- Service Account User
- Storage Admin
- Compute Network Admin

## Environment variables

Set the following environment variables before running the `cluster.sh` script:

| Variable | Description |
|-----|---------|
|**DOCKER_REGISTRY** | Specifies the Docker registry, for example: `docker.io`. |
|**DOCKER_PUSH_REPOSITORY** | Specifies the Docker repository hostname, for example: `docker.io/{anyrepository}`. |
|**DOCKER_PUSH_DIRECTORY** | Specifies the Docker top-level directory, for example: `/home/${USER}/go/src/github.com/kyma-project`.|
|**CLOUDSDK_CORE_PROJECT** | Indicates the GCP project for all GCP resources used during the script execution. |
|**CLOUDSDK_COMPUTE_REGION** | Specifies the GCP compute region, for example `europe-west3`. |
|**CLOUDSDK_COMPUTE_ZONE** | Specifies the GCP compute zone, for example `europe-west3-a`. |
|**GOOGLE_APPLICATION_CREDENTIALS** | Provides the GCP service account key file path, for example: `/etc/credentials/sa-gke-kyma-integration/service-account.json`. |
|**INPUT_CLUSTER_NAME** | Provides a name for the new cluster. |
|**DOCKER_IN_DOCKER_ENABLED** | Specifies the cluster mode. Set this value to `true` for the production mode, or to `false` for the development mode. |

## Script arguments

Set these arguments while running the `cluster.sh` script:

| Name | Required |  Description |
|-----|---------|------------|
|**action** | YES | Specifies the action executed on the Kyma cluster. The possible values are `create` or `delete`. |
|**cluster-grade** | YES | Indicates the cluster grade of the Kyma cluster. The possible values are `production` or `development`. Set the `development` value if you use the cluster for testing purposes. |

>**CAUTION:** If you don't specify these arguments, you will receive an error while running the script.

## Development workflow

>**NOTE:** If you use Dockerhub to push the Kyma Installer image, log in to Docker before you start the `cluster.sh` script.

Follow these steps to run the [`cluster.sh`](../performance-cluster/cluster.sh) script in the development mode:

1. Set up these environment variables:

- **REPO_OWNER** which is the repository owner.
- **REPO_NAME** which is the repository name.

2. Create the development cluster and set the required arguments:
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
5. Delete Kyma and remove the GKE cluster:
```bash
./cluster.sh --action delete
```
