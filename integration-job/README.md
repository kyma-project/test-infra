# Kyma Integration Job

## Overview

This folder contains the details for Kyma integration job. It first creates a VM instance on Google Cloud and installs the dependencies such as Docker and Minikube. Then, it deploys Kyma on Minikube and runs the integration tests.

## Prerequisite

Install the following tools:

    - [gcloud](https://cloud.google.com/sdk/gcloud/)
    - [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)  

Make sure you authenticate gcloud (see [here](https://cloud.google.com/sdk/docs/authorizing)) and kubectl is configured to work with your Prow cluster (see [here](https://cloud.google.com/kubernetes-engine/docs/how-to/cluster-access-for-kubectl)).

### Create a Google Cloud Service Account

To be able to create and interact with a VM, Kyma integration jobs needs to be authenticated using a Service Account.

Create a Service Account and grant the following roles:

    - Service Account User
    - Compute Admin
    - Compute OS Admin Login

```gcloud iam service-accounts create [SA-NAME] --display-name "[SA-DISPLAY-NAME]"```

```gcloud projects add-iam-policy-binding [PROJECT-ID] --member serviceAccount:[SA-NAME]@[PROJECT-ID].iam.gserviceaccount.com --role roles/iam.serviceAccountUser```

```gcloud projects add-iam-policy-binding [PROJECT-ID] --member serviceAccount:[SA-NAME]@[PROJECT-ID].iam.gserviceaccount.com --role roles/compute.instanceAdmin```

```gcloud projects add-iam-policy-binding [PROJECT-ID] --member serviceAccount:[SA-NAME]@[PROJECT-ID].iam.gserviceaccount.com --role roles/compute.osAdminLogin```

Generate key.json for the Service Account:

```gcloud iam service-accounts keys create ~/key.json --iam-account [SA-NAME]@[PROJECT-ID].iam.gserviceaccount.com```

### Create a Secret

Create a Secret on the Prow cluster based on this key.json:

```kubectl create secret generic gc-service-account --from-file=key.json={path-to-your-file}.json```