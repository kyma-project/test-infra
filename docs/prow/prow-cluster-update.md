# Prow cluster update

Updating a Prow cluster requires an improved Prow version. The Kubernetes Prow instance gets updated via a shell script. The shell script offers only a short list of the last pushed container tags and as a result, limits the versions to choose from. To cherry-pick updates, monitor [Prow announcements](https://github.com/kubernetes/test-infra/blob/master/prow/ANNOUNCEMENTS.md) to see when fixes or important changes are merged into the Kubernetes repository. This document describes how to update a Prow cluster using a cherry-picked Prow version.

## Update process

To update a Prow cluster follow these steps:

1. Follow [this](./prow-installation-on-forks.md) document for details to set up a Prow cluster.
2. Go to the [`kubernetes/test-infra`](https://github.com/kubernetes/test-infra/) project and select a commit with the desired update for the Prow cluster. For example, use [`2c8e0dbb96b4c1a86d42275dfbed5474a6d05def`](https://github.com/kubernetes/test-infra/commit/2c8e0dbb96b4c1a86d42275dfbed5474a6d05def).
3. Open both [`starter.yaml`](../../prow/cluster/starter.yaml) in the current project and [`starter.yaml`](https://github.com/kubernetes/test-infra/blob/2c8e0dbb96b4c1a86d42275dfbed5474a6d05def/prow/cluster/starter.yaml) in the Kubernetes project and copy new tags for these containers:
    * gcr.io/k8s-prow/hook
    * gcr.io/k8s-prow/plank
    * gcr.io/k8s-prow/sinker
    * gcr.io/k8s-prow/deck
    * gcr.io/k8s-prow/horologium
    * gcr.io/k8s-prow/tide
4. Copy the commit ID into a comment at the top of the file to keep track of the current release used for the deployments.
5. Check with your preferred diff tool for other meaningful changes (e.g. stability update in a prow component or additional configurations on existing components) that need to be taken into account and copied over.
6. Open both [`config.yaml`](../../prow/config.yaml) in the current project and [`config.yaml`](https://github.com/kubernetes/test-infra/blob/2c8e0dbb96b4c1a86d42275dfbed5474a6d05def/prow/config.yaml) in the Kubernetes project and copy new tags for these containers:
    * gcr.io/k8s-prow/clonerefs
    * gcr.io/k8s-prow/initupload
    * gcr.io/k8s-prow/entrypoint
    * gcr.io/k8s-prow/sidecar
7. Check with your preferred diff tool for other meaningful changes (e.g. stability update in a prow component or additional configurations on existing components) that need to be taken into account and copied over.
8. Run this command to update Prow deployments:
    ```bash
    kubectl apply -f prow/cluster/starter.yaml
    ```
9. Use the `update-config.sh {file_path}` script to apply the Prow configuration on a cluster. Run the following command:
   ```
   ./update-config.sh ../prow/config.yaml
   ```
10. Make sure that the update was successful. For example, run `kubectl get pods` to check if it doesn't show errors on the updated test cluster and the dashboard is still reachable.
11. Create a pull request.

## Troubleshooting

If after the update pods don't start, use the previous commit to bring back the previous deployments' version. Follow these steps:

1. Run the following command to deploy the previously running containers back into the cluster.
    ```bash
    kubectl apply -f prow/cluster/starter.yaml
    ```
2. Use the following command to bring the previous config back:
    ```
    ./update-config.sh ../prow/config.yaml
    ```
3. Use `kubectl get pods` to monitor if the pods start.
