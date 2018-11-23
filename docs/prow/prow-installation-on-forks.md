# Prow Installation on Forks

This instruction provides the steps required to deploy your own Prow on a forked repository for test and development purposes.

> **NOTE:** The following instructions assume that you are signed in to the Google Cloud project with administrative rights and that you have the `$GOPATH` already set.

## Prerequisites

Install the following tools:

- Kubernetes 1.10+ on Google Kubernetes Engine (GKE)
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) to communicate with Kubernetes
- [gcloud](https://cloud.google.com/sdk/gcloud/) to communicate with Google Cloud Platform (GCP)
- OpenSSL

## Provision a cluster

1. Export these variables:

```

export PROJECT={project-name}
export CLUSTER_NAME={cluster-name}
export ZONE={zone-name}

```

2. When you communicate for the first time with Google Cloud, set the context to your Google Cloud project. Run this command:

```
gcloud config set project $PROJECT
```

3. Run the [`provision-cluster.sh`](../../development/provision-cluster.sh) script or follow [this](https://github.com/kubernetes/test-infra/blob/master/prow/getting_started_deploy.md#create-the-cluster) instruction to provision a new cluster on GKE. Make sure that kubectl points to the correct cluster. For GKE, run the following command:

```
gcloud container clusters get-credentials $CLUSTER_NAME --zone=$ZONE --project=$PROJECT
```

## Create a bot account

Create a separate GitHub account which serves as a bot account that triggers the Prow comments that you enter in the pull request. If the Prow bot account is the same as the account that creates a job-triggering comment, the job is not triggered.

Add the bot account to the [collaborators](https://help.github.com/articles/adding-outside-collaborators-to-repositories-in-your-organization/) on your forked repository and set it with push access rights. The bot account must accept your invitation.

## Set an access token

Set an OAuth2 token that has the read and write access to the bot account.

To generate a new token, go to the **Settings** tab of a given GitHub account and click **Developer Settings**. Choose **Personal Access Token** and **Generate New Token**.
In the new window, select all scopes and click **Generate token**.

You can set the token either as an environment variable named `OAUTH` or provide it during the installation process when prompted.

## Create Secrets

For the purpose of the installation, you must have a set of service accounts and secret files created on Google Cloud Storage (GCS).

> **NOTE:** For details, see the [Prow Secrets Management](./prow-secrets-management.md) document that explains step by step how to create all required GCS resources.

1. Create two buckets on GCS, one for storing Secrets and the second for storing logs.

> **NOTE:** The bucket for storing logs is used in Prow by the Plank component. This reference is defined in the `config.yaml` file.

2. Create the following service accounts, role bindings, and private keys. Encrypt them using Key Management Service (KMS), and upload them to your Secret storage bucket:

- **sa-gke-kyma-integration** with roles that allow the account to create Kubernetes clusters:
  - Compute Network Admin (`roles/compute.networkAdmin`)
  - Kubernetes Engine Admin (`roles/container.admin`)
  - Kubernetes Engine Cluster Admin (`roles/container.clusterAdmin`)
  - DNS Administrator (`roles/dns.admin`)
  - Service Account User (`roles/iam.serviceAccountUser`)
  - Storage Admin (`roles/storage.admin`)
- **sa-vm-kyma-integration** with roles that allow the account to provision virtual machines:
  - Compute Instance Admin (beta) (`roles/compute.instanceAdmin`)
  - Compute OS Admin Login (`roles/compute.osAdminLogin`)
  - Service Account User (`roles/iam.serviceAccountUser`)
- **sa-gcs-plank** with the role that allows the account to store objects in a bucket:
  - Storage Object Admin (`roles/storage.objectAdmin`)
- **sa-gcr-push** with the role that allows the account to push images to Google Container Repository:
  - Storage Admin `roles/storage.admin`

## Install Prow

Follow these steps to install Prow:

1. Export these environment variables:

- **BUCKET_NAME** is a GCS bucket in the Google Cloud project that stores Prow Secrets.
- **KEYRING_NAME** is the KMS key ring.
- **ENCRYPTION_KEY_NAME** is the key name in the key ring that is used for data encryption.

The account files are encrypted with the **ENCRYPTION_KEY_NAME** key from **KEYRING_NAME** and are stored in **BUCKET_NAME**.

2. Go to the `development` folder and run the following script to start the installation process:

```bash
./install-prow.sh
```

> **NOTE:** The scripts prompts you to enter your OAuth2 token.

This script performs the following steps to install Prow:

- Deploy the NGINX Ingress Controller.
- Create a ClusterRoleBinding.
- Create a HMAC token used for GitHub webhooks.
- Create Secrets for HMAC and OAuth2 used by Prow.
- Deploy Prow components using the `starter.yaml` file from the `prow/cluster` directory.
- Add annotations for the Prow Ingress to make it work with the NGINX Ingress Controller.

## Verify the Installation

Verify if the Prow installation was successful.

1. Check if all Pods are up and running:

   ```
   kubectl get pods
   ```

2. Check if the Deck is accessible from outside of the cluster:

   ```
   kubectl get ingress ing
   ```

   Copy the address of the ingress `ing` and open it in a browser to display the Prow status on the dashboard.

## Configure the webhook

After Prow installs successfully, you must [configure the webhook](https://support.hockeyapp.net/kb/third-party-bug-trackers-services-and-webhooks/how-to-set-up-a-webhook-in-github) to enable the GitHub repository to send events to Prow.

## Configure Prow

When you use the [`install-prow.sh`](../../development/provision-cluster.sh) script to install Prow on your cluster, the list of plugins and configuration is empty. You can configure Prow by specifying the `config.yaml` and `plugins.yaml` files, and adding job definitions to the `jobs` directory.

### The config.yaml file

The `config.yaml` file contains the basic Prow configuration. When you create a particular ProwJob, it uses the Preset definitions from this file. See the example of such a file [here](../../prow/config.yaml).

For more details, see the [Kubernetes documentation](https://github.com/kubernetes/test-infra/blob/master/prow/getting_started_deploy.md#add-more-jobs-by-modifying-configyaml).

### The plugins.yaml file

The `plugins.yaml` file contains the list of [plugins](https://status.build.kyma-project.io/plugins) you enable on a given repository. See the example of such a file [here](../../prow/plugins.yaml).

For more details, see the [Kubernetes documentation](https://github.com/kubernetes/test-infra/blob/master/prow/getting_started_deploy.md#enable-some-plugins-by-modifying-pluginsyaml).

### The job configuration file

You can define a test presubmit job for a component. However, remember to adjust its definition in the `yaml` file to point to your forked repository instead of the original repository.

For details on how to define a presubmit job, see the [Migration Guide](./migration-guide.md#create-a-presubmit-job).

### Verify the configuration

To check if the `plugins.yaml`, `config.yaml`, and jobs configuration files are correct, run the `validate-config.sh {plugins_file_path} {config_file_path} {jobs_dir_path}` script. For example, run:

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

After you complete the required configuration, you can test the uploaded plugins and configuration. You can also create your own job pipeline and test it against the forked repository.

### Cleanup

To clean up everything created by the installation script, run the removal script:

```
./remove-prow.sh
```
