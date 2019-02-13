# Presets

This document contains the list of all Presets available in the [`config.yaml`](../../prow/config.yaml) file. Use them to define ProwJobs for your components.

| Name                               | Description                                                                                                                                                     |
| ---------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **preset-dind-enabled**            | It allows Docker to run in your job.                                                                                                                        |
| **preset-sa-gcr-push**             | It injects credentials for pushing images to Google Cloud Registry (GCR).                                                                                             |
| **preset-docker-push-repository**  | It provides the environment variable with the location of the Docker repository.                                                                                 |
| **preset-build-pr**                | It provides the environment variable with the location of the directory in the Docker repository for storing images. It also sets the **BUILD_TYPE** variable to `pr`. |
| **preset-build-master**            | It is similar to the **preset-build-pr** Preset, but the **BUILD_TYPE** variable is set to `master`.                                                            |
| **preset-build-release**           | It is similar to the **preset-build-pr** Preset, but the **BUILD_TYPE** variable is set to `release`.                                                           |
| **preset-gc-project-env**          | It provides the environment variable with the Google Cloud Platform (GCP) project name.                                                                                              |
| **preset-gc-compute-envs**         | It provides environment variables with the GCP compute zone and the GCP compute region.                                                                   |
| **preset-sa-vm-kyma-integration**  | It injects credentials for the service account to run integration tests on virtual machines (VMs).                                                              |
| **preset-sa-gke-kyma-integration** | It injects credentials for the service account to run integration tests on a Google Cloud Engine (GKE) cluster.                                                 |
| **preset-bot-npm-token**           | It provides an environment variable with a token for publishing npm packages.
| **preset-sa-kyma-artifacts** | It sets up the service account that has write permissions to the Kyma's artifacts bucket.                     |
| **preset-docker-push-repository-gke-integration** | It provides the environment variable with the location of the directory in the GCR repository for storing temporary Docker images for the Kyma Installer.                     |
| **preset-docker-push-repository-test-infra** | It defines the environment variable with the location of the directory in the GCR repository for storing Docker images from the `test-infra` repository.                    |
| **preset-docker-push-repository-incubator** | It defines the environment variable with the location of the directory in the GCR repository for storing Docker images from all repositories under the `kyma-incubator` organization.                     |
| **preset-build-console-master** | It defines the environment variable with the location of the directory in the Docker repository for storing Docker images from the `console` repository. It also sets the **BUILD_TYPE** variable to `master`.                  |
| **preset-bot-github-token** | It sets the environment variable with the GitHub token for the bot account.                     |
| **preset-bot-github-identity** | It sets the environment variables for the name and email of the bot account.               |
| **preset-bot-github-ssh** | It connects the ssh key of the bot account to your job and sets the value with the path to this key.                    |
| **preset-kyma-artifacts-bucket** | It defines the environment variable for the Kyma's artifact bucket.                     |
| **preset-stability-checker-slack-notifications** | It defines a webhook URL and a client token required for the Slack integration.                 |
| **preset-sap-slack-bot-token** | It sets the environment variable for the Slack token of the bot account in the SAP CX workspace. Find more information [here](https://api.slack.com/docs/token-types#bot).|
| **preset-kyma-ondemands** | It defines the environment variable for the Kyma's ondemand artifacts bucket. |