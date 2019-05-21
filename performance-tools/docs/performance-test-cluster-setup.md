# Performance test setup

This document describes how to install the load generating cluster.

## Prerequisites

- Create a GKE account with `owner` rights.
- Prepare a domain name that you will use.

## Installation

1. [Install Kyma](https://kyma-project.io/docs/root/kyma/#installation-install-kyma-on-a-cluster) on the cluster. Kyma comes with authentication and authorization, as well as Grafana.

2. [Install InfluxDB](https://console.cloud.google.com/marketplace/details/google/influxdb?q=influxdb) to store test results. The supported version of InfluxDB is 1.6.

3. Use Grafana to visualize test results and configure it to read metrics from InfluxDB.

4. Configure InfluxDB:
   * Log in to Influx:
     ```bash
     influx -username {username}  -password {password}
     ```
   * Create a USER with both read and write rights.
     ```bash
      CREATE USER readUser with PASSWORD {password}
      CREATE USER writeUser with PASSWORD {password}
     ```
   * Create a database:
     ```bash
     CREATE DATABASE "database"
     ```
   * Grant appropriate rights to users:
      ```bash
      GRANT ALL ON "database" TO "writeUser"
      GRANT READ ON "database" TO "readUser"
      ```

5. Configure the GKE cluster:
  * Export the required variables into shell:
      ```bash
      export GCLOUD_PROJECT_NAME="project-name"
      export GCLOUD_COMPUTE_ZONE="compute-zone"
      export BUCKET_NAME="bucket-name"
      export KEYRING_NAME="keyring-name"
      export ENCRYPTION_KEY_NAME="enc-key-name"
      export LOAD_GEN_CLUSTER="load-gen-cluster"
      export LOAD_GEN_NAMESPACE="load-gen-ns"
      ```
  * Set up the Google Cloud Project where the GKE cluster will be created:
    ```bash
    gcloud config set project $GCLOUD_PROJECT_NAME
    ```
  * Create a bucket:
    ```bash
    gsutil mb -p $GCLOUD_PROJECT_NAME gs://$BUCKET_NAME/
    ```
  * Create a keyring and keys for kms:
    ```bash
    gcloud kms keyrings create $KEYRING_NAME --location "global"
    gcloud kms keys create $ENCRYPTION_KEY_NAME --location "global" --keyring $KEYRING_NAME --purpose encryption
    ```
  * Create a service account:
    ```bash
    export SA_NAME="sa-name"
    export SA_DISPLAY_NAME=$SA_NAME
    export SECRET_FILE="sa.json"

    gcloud iam service-accounts create $SA_NAME --display-name $SA_DISPLAY_NAME
    gcloud iam service-accounts keys create $SECRET_FILE --iam-account=$SA_NAME@$GCLOUD_PROJECT_NAME.iam.gserviceaccount.com
    ```
    Save the service account credentials into the GCP bucket:
    ```bash
    gcloud kms encrypt --location global --keyring $KEYRING_NAME --key $ENCRYPTION_KEY_NAME --plaintext-file $SECRET_FILE --ciphertext-file $SECRET_FILE.encrypted
    gsutil cp $SECRET_FILE.encrypted gs://$BUCKET_NAME/
    ```
  * Add the required roles to the service account:
    ```bash
      for role in "roles/container.admin" "roles/container.clusterAdmin" "roles/serviceaccounts.serviceAccountUser", "roles/storage.storageAdmin"; do
          echo $role
          gcloud projects add-iam-policy-binding $GCLOUD_PROJECT_NAME  --member=serviceAccount:$SA_NAME@$GCLOUD_PROJECT_NAME.iam.gserviceaccount.com --role=$role
        done
    ```
  * Create a Namespace for running the performance test job:
    ```bash
    gcloud container clusters get-credentials $LOAD_GEN_CLUSTER --zone=$GCLOUD_COMPUTE_ZONE --project=$GCLOUD_PROJECT_NAME

    kubectl create ns $LOAD_GEN_NAMESPACE

    kubectl label ns $LOAD_GEN_NAMESPACE env=true
    ```
  * Create Secrets for the service account:
    ```bash
    kubectl create secret generic $SA_NAME --from-file=./sa.json -n $LOAD_GEN_NAMESPACE
    ```

  * Create Secrets for InfluxDB `readUser` and `writeUser`:
    ```bash
    kubectl create secret generic `k6-secrets` --from-file=./writeUser --from-file=./writeUser_pass --from-file=./database -n $LOAD_GEN_NAMESPACE
    ```
