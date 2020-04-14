# Production Cluster Configuration

## Overview

This instruction provides the steps required to deploy a production cluster for Prow.
>**NOTE**: This Prow installation is compatible with the [`4faaf685958cd79ea5b5a376fadabd8a9d1b4123`](https://github.com/kubernetes/test-infra/commit/4faaf685958cd79ea5b5a376fadabd8a9d1b4123) revision in the `kubernetes/test-infra` repository.

## Prerequisites

Use the following tools and configuration:

- Kubernetes 1.10+ on Google Kubernetes Engine (GKE)
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) to communicate with Kubernetes
- [gcloud](https://cloud.google.com/sdk/gcloud/) to communicate with Google Cloud Platform (GCP)
- The `kyma-bot` GitHub account
- [Kubernetes cluster](./prow-installation-on-forks.md#provision-a-main-prow-cluster)
- Secrets in the Kubernetes cluster:
  - `hmac-token` which is a Prow HMAC token used to validate GitHub webhooks
  - `oauth-token` which is a GitHub token with read and write access to the `kyma-bot` account
  - `sap-slack-bot-token` which is a token for publishing messages in the SAP CX workspace. Find more information [here](https://api.slack.com/docs/token-types#bot).
- Two buckets on Google Cloud Storage (GCS), one for storing Secrets and the second for storing logs
- GCP configuration that includes:
  - A [global static IP address](https://cloud.google.com/compute/docs/ip-addresses/reserve-static-external-ip-address) with the `kyma-prow-status` name
  - A [DNS registry](https://cloud.google.com/dns/docs/quickstart#create_a_managed_public_zone) for the `status.build.kyma-project.io` domain that points to the `kyma-prow-status` address


## Installation

1. Prepare the workload cluster:

  ```bash
    export WORKLOAD_CLUSTER_NAME=kyma-prow-workload
    export ZONE=europe-west3-a
    export PROJECT=sap-kyma-prow

    ### In GKE get KUBECONFIG for cluster kyma-prow-workload
    gcloud container clusters get-credentials $WORKLOAD_CLUSTER_NAME --zone=$ZONE --project=$PROJECT

    ./set-up-workload-cluster.sh
  ```

  This script performs the following steps:
  - Creates a ClusterRoleBinding to provide access to the Prow cluster. This way it enables running and monitoring jobs on the workload cluster.
  - Creates Kubernetes Secrets resources from secrets fetched from the GCP bucket.

2. Set the context to your Google Cloud project.

    Export the **PROJECT** variable and run this command:

  ```bash
    gcloud config set project $PROJECT
  ```

3. Make sure that kubectl points to the Prow main cluster.

  Export these variables:

  ```bash
    export CLUSTER_NAME=kyma-prow
    export ZONE=europe-west3-a
    export PROJECT=sap-kyma-prow
  ```

   For GKE, run the following command:

  ```bash
    gcloud container clusters get-credentials $CLUSTER_NAME --zone=$ZONE --project=$PROJECT
  ```

4. Export these environment variables:

  ```bash
    export BUCKET_NAME=kyma-prow-secrets
    export KEYRING_NAME=kyma-prow
    export ENCRYPTION_KEY_NAME=kyma-prow-encryption
    export GOPATH=$GOPATH ### Ensure GOPATH is set
  ```
where:
   - **BUCKET_NAME** is a GCS bucket in the Google Cloud project that stores Prow Secrets.
   - **KEYRING_NAME** is the KMS key ring.
   - **ENCRYPTION_KEY_NAME** is the key name in the key ring that is used for data encryption.

5. Run the following script to create a Kubernetes Secret resource in the main Prow cluster. This way the main Prow cluster can access the workload cluster:

  ```bash
    ./create-secrets-for-workload-cluster.sh
  ```

>**NOTE:** Create the workload cluster first and make sure the **local** kubeconfig for the Prow admin contains the context for this cluster. Point the **current** kubeconfig to the main Prow cluster.

6. Run the following script to start the installation process:

  ```bash
    ./scripts/install-prow.sh
  ```

   The installation script performs the following steps to install Prow:

   - Deploys the NGINX Ingress Controller
   - Creates a ClusterRoleBinding
   - Deploys Prow components with the `a202e595a33ac92ab503f913f2d710efabd3de21`revision
   - Deploys the Cert Manager
   - Deploys secure Ingress
   - Deploys the [Prow Addons Controller Manager](../../development/prow-addons-ctrl-manager/README.md)
   - Removes insecure Ingress

7. Verify the installation.

   To check if the installation is successful, perform the following steps:

   - Check if all Pods are up and running:
     `kubeclt get pods`
   - Check if the Deck is accessible from outside of the cluster:
     `kubectl get ingress tls-ing`
   - Copy the address of the `tls-ing` Ingress and open it in a browser to display the Prow status on the dashboard.

## Configure Prow

When you use the [`install-prow.sh`](../../prow/install-prow.sh) script to install Prow on your cluster, the list of plugins and configuration is empty. You can configure Prow by specifying the `config.yaml` and `plugins.yaml` files, and adding job definitions to the `jobs` directory.

### The config.yaml file

The `config.yaml` file contains the basic Prow configuration. When you create a particular Prow job, it uses the Preset definitions from this file. See the example of such a file [here](../../prow/config.yaml).

For more details, see the [Kubernetes documentation](https://github.com/kubernetes/test-infra/blob/master/prow/getting_started_deploy.md#add-more-jobs-by-modifying-configyaml).

### The plugins.yaml file

The `plugins.yaml` file contains the list of [plugins](https://status.build.kyma-project.io/plugins) you enable on a given repository. See the example of such a file [here](../../prow/plugins.yaml).

For more details, see the [Kubernetes documentation](https://github.com/kubernetes/test-infra/blob/master/prow/getting_started_deploy.md#enable-some-plugins-by-modifying-pluginsyaml).

### The jobs directory

The `jobs` directory contains the Prow jobs configuration. See the example of such a file [here](../../prow/jobs).

For more details, see the [Kubernetes documentation](https://github.com/kubernetes/test-infra/blob/master/prow/getting_started_deploy.md#add-more-jobs-by-modifying-configyaml).

### Verify the configuration

To check if the `plugins.yaml`, `config.yaml`, and `jobs` configuration files are correct, run the `validate-config.sh {plugins_file_path} {config_file_path} {jobs_dir_path}` script. For example, run:

```bash
  ./validate-config.sh ../prow/plugins.yaml ../prow/config.yaml ../prow/jobs
```

### Upload the configuration on a cluster

If the files are configured correctly, upload the files on a cluster.

1. Use the `update-plugins.sh {file_path}` script to apply plugin changes on a cluster.

   ```bash
   ./update-plugins.sh ../prow/plugins.yaml
   ```

2. Use the `update-config.sh {file_path}` script to apply Prow configuration on a cluster.

   ```bash
   ./update-config.sh ../prow/config.yaml
   ```

3. Use the `update-jobs.sh {jobs_dir_path}` script to apply jobs configuration on a cluster.

   ```bash
   ./update-jobs.sh ../prow/jobs
   ```

After you complete the required configuration, you can test the uploaded plugins and configurations. You can also create your own job pipeline and test it against the forked repository.

### Cleanup

To clean up everything created by the installation script, run the removal script:

```bash
./remove-prow.sh
```
