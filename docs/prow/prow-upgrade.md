# Prow cluster update

In order to get improved/newer versions of prow we need to upgrade prow. The kubernetes prow instance is getting updated via a shell script that is very limiting in the versions that you can choose from and is probably not good enough for cherrypicking updates we like. In the meantime though we can't not update the cluster and will have to describe a way to do updates. This document aims to provide a description of such an upgrade.

## Update process

To update a Prow cluster follow these steps:

1. Follow [this](./prow-installation-on-forks.md) prow-installation-on-forks.md) document for details to set up a Prow cluster.
2. In the [kubernetes/test-infra](https://github.com/kubernetes/test-infra/) project select the commit that is going to be your the one the prow cluster should be upgraded to. E.g. [2c8e0dbb96b4c1a86d42275dfbed5474a6d05def](https://github.com/kubernetes/test-infra/commit/2c8e0dbb96b4c1a86d42275dfbed5474a6d05def).
3. Open both [`starter.yaml`](../../prow/cluster/starter.yaml) in the current project and [`starter.yaml`](https://github.com/kubernetes/test-infra/blob/2c8e0dbb96b4c1a86d42275dfbed5474a6d05def/prow/cluster/starter.yaml) in the Kubernetes project and copy new lables for these containers:
    * gcr.io/k8s-prow/hook
    * gcr.io/k8s-prow/plank
    * gcr.io/k8s-prow/sinker
    * gcr.io/k8s-prow/deck
    * gcr.io/k8s-prow/horologium
    * gcr.io/k8s-prow/tide
4. Copy commit id into comment on top of the file to keep track of the current release used for the deployments (due to prow not having any releases)
5. Check with a diff tool for other meaningful changes that need to be taken into account and copied over.
6. Open both [`config.yaml`](../../prow/config.yaml) in the current project and [`config.yaml`](https://github.com/kubernetes/test-infra/blob/2c8e0dbb96b4c1a86d42275dfbed5474a6d05def/prow/config.yaml) in the Kubernetes project and copy new labels for these containers:
    * gcr.io/k8s-prow/clonerefs
    * gcr.io/k8s-prow/initupload
    * gcr.io/k8s-prow/entrypoint
    * gcr.io/k8s-prow/sidecar
7. Use a diff tool to find other meaningful changes and copy the items you need.
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

1. In case something goes wrong with the upgrade and pods are not starting anymore, check out the old commit before the update happened and change the deployments back to what they were via:
    ```bash
    kubectl apply -f prow/cluster/starter.yaml
    ```
2. Use the following command to bring the previous config back:
    ```
    ./update-config.sh ../prow/config.yaml
    ```
3. Use `kubectl get pods` to monitor if the pods start.
