# Performance test setup

## Installation of Load Generating Cluster.

### Pre-requisites

We have used GKE for installtinng the load generator cluster. Following are the prerequisites for the same.

1. Create GKE account with `onwer` rights
2. Have domain name to be used.

### Installation

#### 1. Installation of the cluster
We are using kyma for the installation as it provides authentication and authorization. Additionally we can leverage grafana which comes with Kyma. To install the cluster follow the steps documented [here](https://kyma-project.io/docs/root/kyma/#installation-install-kyma-on-a-cluster).


#### 2. Installation of InfluxDB
k6 uses InfluxDB to store results and Grafana could be used to visualize the test results. To install InfluxDB we used the application available on [GCP market place](https://console.cloud.google.com/marketplace/details/google/influxdb?q=influxdb). The version of influxDB has been backported to [1.6](gcr.io/cloud-marketplace/google/influxdb@sha256:23d3f92f3f375a7e37ee4e54e739a068e9cf80a570ffecce60b97076c15855b6`), since 1.7 has issues with random timeouts/freeze as mentioned [here](https://github.com/influxdata/influxdb/issues/12731) 



#### 3. Configuration

Following are the configurations required:
1. InfluxDB
   * Login into influx:
     ```bash
     influx -username <>  -password <>
     ```
   * Create USER to readUser and writeUser
     ```bash
      CREATE USER readUser with PASSWORD <>
      CREATE USER writeUser with PASSWORD <>
     ```
   * Create database
     ```bash
     CREATE DATABASE "database"
     ```
   * GRANT appropriate rights to users
      ```bash
      GRANT ALL ON "databse" TO "writeUser"
      GRANT READ ON "database" TO "readUser"
      ```
2. Grafana
  Configure Grafana to use `readUser` to read the metrics from influxDB.
3. GKE
  * Export the required variables into shell
      ```bash
      export GCLOUD_PROJECT_NAME="project-name"
      export GCLOUD_COMPUTE_ZONE="compute-zone"
      export BUCKET_NAME="bucket-name"
      export KEYRING_NAME="keyring-name"
      export ENCRYPTION_KEY_NAME="enc-key-name"
      export LOAD_GEN_CLUSTER="load-gen-cluster"
      export LOAD_GEN_NAMESPACE="load-gen-ns"
      ```
  * Set the project
    ```bash
    gcloud config set project $GCLOUD_PROJECT_NAME
    ```
  * Create Bucket
    ```bash
    gsutil mb -p $GCLOUD_PROJECT_NAME gs://$BUCKET_NAME/
    ```
  * Create keyring and keys for kms
    ```bash
    gcloud kms keyrings create $KEYRING_NAME --location "global"
    gcloud kms keys create $ENCRYPTION_KEY_NAME --location "global" --keyring $KEYRING_NAME --purpose encryption
    ```
  * Service Account
    ```bash
    export SA_NAME="sa-name"
    export SA_DISPLAY_NAME=$SA_NAME
    export SECRET_FILE="sa.json"

    gcloud iam service-accounts create $SA_NAME --display-name $SA_DISPLAY_NAME
    gcloud iam service-accounts keys create $SECRET_FILE --iam-account=$SA_NAME@$GCLOUD_PROJECT_NAME.iam.gserviceaccount.com
    ```
    Save the service account credentials into GCP bucket
    ```bash
    gcloud kms encrypt --location global --keyring $KEYRING_NAME --key $ENCRYPTION_KEY_NAME --plaintext-file $SECRET_FILE --ciphertext-file $SECRET_FILE.encrypted
    gsutil cp $SECRET_FILE.encrypted gs://$BUCKET_NAME/
    ```
    Add the required roles to the service account
    ```bash
      for role in "roles/container.admin" "roles/container.clusterAdmin" "roles/serviceaccounts.serviceAccountUser", "roles/storage.storageAdmin"; do
          echo $role
          gcloud projects add-iam-policy-binding $GCLOUD_PROJECT_NAME  --member=serviceAccount:$SA_NAME@$GCLOUD_PROJECT_NAME.iam.gserviceaccount.com --role=$role
        done
    ```
  * Create the namespace for running the performance test job
    ```bash
    gcloud container clusters get-credentials $LOAD_GEN_CLUSTER --zone=$GCLOUD_COMPUTE_ZONE --project=$GCLOUD_PROJECT_NAME

    kubectl create ns $LOAD_GEN_NAMESPACE

    kubectl label ns $LOAD_GEN_NAMESPACE env=true
    ```
  * Create various secrets
    create secret for service account
    ```bash
    kubectl create secret generic $SA_NAME --from-file=./sa.json -n $LOAD_GEN_NAMESPACE
    ```

    create secret for influxDB `readUser`  and `readWriter`
    ```bash
    kubectl create secret generic `k6-secrets` --from-file=./writeUser --from-file=./writeUser_pass --from-file=./database -n $LOAD_GEN_NAMESPACE