# Prow Cluster Monitoring Setup

This document describes how to install and manage Prow cluster monitoring that is available at `https://monitoring.build.kyma-project.io`. This document also describes how to create and manage Grafana dashboards.

## Prerequisites

Install the following tools:

- Helm v2.11.0
- kubectl

## Configure Slack for failure notifications

Follow these steps:

1. Create a Slack channel and an [Incoming Webhook](https://api.slack.com/incoming-webhooks) for this channel. Copy the resulting Webhook URL.

2. Replace `{SLACK_URL}` in [values.yaml](./../../prow/cluster/resources/monitoring/values.yaml) with the Weebhook URL and `{SLACK_CHANNEL}` with the channel name.

3. Replace `{PROW_SLACK_URL}` with prow channel Weebhook URL and `{PROW_SLACK_CHANNEL}` with the channel name in [prow-slack-config.yaml](./../../prow/cluster/resources/monitoring/prow-slack-config.yaml).

## Provision a monitoring chart

Follow these steps:

1. Make sure that kubectl points to the correct cluster.
   
   ```bash
   gcloud container clusters get-credentials {clusterName} --zone={zoneName} --project={projectName}
   ```

2. Go to the [`prow/cluster`](../../prow/cluster) directory.

3. Download dependencies:
   
   ```bash
   helm dependency build resources/monitoring
   ```

4. Install the monitoring chart:

   ```bash
   helm install --name {releaseName} --namespace {namespaceName} resources/monitoring -f values.yaml,prow-prometheus-rules.yaml,prow-slack-config.yaml
   ```

5. Open the Grafana dashboard.
   
   Grafana dashboard is available at `https://monitoring.build.kyma-project.io`. It can take some time till the dashboard is accessible.

## Authenticate to Grafana

By default, Grafana dashboards are visible for anonymous users with the read-only access. Only authenticated users are able to create and edit dashboards. To sign in to Grafana, follow this steps:

1. Get the password for the `admin` user from the cluster:

   ```bash
   kubectl -n {namespaceName} get secret {releaseName}-grafana -o jsonpath="{.data.admin-password}" | base64 -D
   ```

2. Go to `https://monitoring.build.kyma-project.io/login`.

3. Provide credentials:

   ```
   Login: admin
   Password: {The value from step 1}
   ```

## Create and edit Grafana dashboards

To create or edit Grafana dashboards you must be signed in. The [official Grafana documentation](http://docs.grafana.org/guides/getting_started/) provides instructions on how to work with the dashboards. The main difference between the official guidelines and the the Kyma implementation of Grafana dashboards is the way in which you store them.

Follow these steps to save the dashboard:

1. Export the dashboard to a JSON format.

2. Save the JSON file under `prow/cluster/resources/monitoring/dashboards/`.

3. Update the Grafana configuration on the cluster.
   
   ```bash
   helm upgrade {releaseName} resources/monitoring --recreate-pods
   ```

   > **NOTE:** `--recreate-pods` is required because the Secret with the Grafana password is regenerated during the upgrade and it needs to be populated to Grafana.
