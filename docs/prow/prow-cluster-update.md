# Prow cluster update

In order to get improved/newer versions of prow we need to update prow. The kubernetes teams' prow instance is getting updated via a shell script. This shell script offers a small list of the last pushed container tags. This could reflect only containers of a single day, due to how many times the kubernetes team is actually running their pipeline. This limits us in the versions we can choose from, as we usually don't go from finding out about a new version, to deploying it, within one day. To cherrypick updates, we will need a different process. By monitoring [prow announcements](https://github.com/kubernetes/test-infra/blob/master/prow/ANNOUNCEMENTS.md) we can see when fixes or important changes are merged into the kubernetes teams' repository and start looking into applying these changes to our prow cluster. This document aims to provide a description of how to perform such an update.

## Update process

To update a Prow cluster follow these steps:

1. Follow [this](./prow-installation-on-forks.md) document for details to set up a Prow cluster.
2. Go to the [`kubernetes/test-infra`](https://github.com/kubernetes/test-infra/) project and select a commit with the desired update for the Prow cluster. For example, use [2c8e0dbb96b4c1a86d42275dfbed5474a6d05def](https://github.com/kubernetes/test-infra/commit/2c8e0dbb96b4c1a86d42275dfbed5474a6d05def).
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
