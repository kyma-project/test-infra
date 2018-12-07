# Production Cluster Configuration

## Overview

This instruction provides the steps required to deploy a production cluster for Prow.

## Prerequisites

Use the following tools and configuration:

- Kubernetes 1.10+ on Google Kubernetes Engine (GKE)
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) to communicate with Kubernetes
- [gcloud](https://cloud.google.com/sdk/gcloud/) to communicate with Google Cloud Platform (GCP)
- The `kyma-bot` GitHub account
- [Kubernetes cluster](./prow-installation-on-forks.md#provision-a-cluster)
- Two Secrets in the Kubernetes cluster:
  - `hmac-token` which is a Prow HMAC token used to validate GitHub webhooks
  - `oauth-token` which is a GitHub token with read and write access to the `kyma-bot` account
- Two buckets on Google Cloud Storage (GCS), one for storing Secrets and the second for storing logs
- GCP configuration that includes:
  - A [global static IP address](https://cloud.google.com/compute/docs/ip-addresses/reserve-static-external-ip-address) with the `prow-production` name
  - A [DNS registry](https://cloud.google.com/dns/docs/quickstart#create_a_managed_public_zone) for the `status.build.kyma-project.io` domain that points to the `prow-production` address

## Installation

1. When you communicate for the first time with Google Cloud, set the context to your Google Cloud project.

   Export the **PROJECT** variable and run this command:

   ```
   gcloud config set project $PROJECT
   ```

2. Make sure that kubectl points to the correct cluster.

   Export these variables:

   ```
   export CLUSTER_NAME=prow-production
   export ZONE=europe-west3-b
   export PROJECT=kyma-project
   ```

   For GKE, run the following command:

   ```
   gcloud container clusters get-credentials $CLUSTER_NAME --zone=$ZONE --project=$PROJECT
   ```

3. Export these environment variables, where:

   - **BUCKET_NAME** is a GCS bucket in the Google Cloud project that stores Prow Secrets.
   - **KEYRING_NAME** is the KMS key ring.
   - **ENCRYPTION_KEY_NAME** is the key name in the key ring that is used for data encryption.

   ```
   export BUCKET_NAME=kyma-prow
   export KEYRING_NAME=kyma-prow
   export ENCRYPTION_KEY_NAME=kyma-prow-encryption
   ```

4. Run the following script to start the installation process:

   ```bash
   ./install-prow.sh
   ```

   The installation script performs the following steps to install Prow:

   - Deploy the NGINX Ingress Controller.
   - Create a ClusterRoleBinding.
   - Deploy Prow components with the `a202e595a33ac92ab503f913f2d710efabd3de21`revision.
   - Deploy the Cert Manager.
   - Deploy secure Ingress.
   - Remove insecure Ingress.

5. Verify the installation.

   To check if the installation is successful, perform the following steps:

   - Check if all Pods are up and running:
     `kubeclt get pods`
   - Check if the Deck is accessible from outside of the cluster:
     `kubectl get ingress tls-ing`
   - Copy the address of the `tls-ing` Ingress and open it in a browser to display the Prow status on the dashboard.

## Configure Prow

When you use the [`install-prow.sh`](../../prow/install-prow.sh) script to install Prow on your cluster, the list of plugins and configuration is empty. You can configure Prow by specifying the `config.yaml` and `plugins.yaml` files, and adding job definitions to the `jobs` directory.

### The config.yaml file

The `config.yaml` file contains the basic Prow configuration. When you create a particular ProwJob, it uses the Preset definitions from this file. See the example of such a file [here](../../prow/config.yaml).

For more details, see the [Kubernetes documentation](https://github.com/kubernetes/test-infra/blob/master/prow/getting_started_deploy.md#add-more-jobs-by-modifying-configyaml).

### The plugins.yaml file

The `plugins.yaml` file contains the list of [plugins](https://status.build.kyma-project.io/plugins) you enable on a given repository. See the example of such a file [here](../../prow/plugins.yaml).

For more details, see the [Kubernetes documentation](https://github.com/kubernetes/test-infra/blob/master/prow/getting_started_deploy.md#enable-some-plugins-by-modifying-pluginsyaml).

### The jobs directory

The `jobs` directory contains the ProwJobs configuration. See the example of such a file [here](../../prow/jobs).

For more details, see the [Kubernetes documentation](https://github.com/kubernetes/test-infra/blob/master/prow/getting_started_deploy.md#add-more-jobs-by-modifying-configyaml).

### Verify the configuration

To check if the `plugins.yaml`, `config.yaml`, and `jobs` configuration files are correct, run the `validate-config.sh {plugins_file_path} {config_file_path} {jobs_dir_path}` script. For example, run:

```
./validate-config.sh ../prow/plugins.yaml ../prow/config.yaml ../prow/jobs
```

### Upload the configuration on a cluster

If the files are configured correctly, upload the files on a cluster.

1. Use the `update-plugins.sh {file_path}` script to apply plugin changes on a cluster.

   ```
   ./update-plugins.sh ../prow/plugins.yaml
   ```

2. Use the `update-config.sh {file_path}` script to apply Prow configuration on a cluster.

   ```
   ./update-config.sh ../prow/config.yaml
   ```

3. Use the `update-jobs.sh {jobs_dir_path}` script to apply jobs configuration on a cluster.

   ```
   ./update-jobs.sh ../prow/jobs
   ```

After you complete the required configuration, you can test the uploaded plugins and configurations. You can also create your own job pipeline and test it against the forked repository.

### Cleanup

To clean up everything created by the installation script, run the removal script:

```
./remove-prow.sh
```
