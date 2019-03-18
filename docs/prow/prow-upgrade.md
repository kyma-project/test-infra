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
4. Copy over new labels for the following containers:
    * gcr.io/k8s-prow/hook
    * gcr.io/k8s-prow/plank
    * gcr.io/k8s-prow/sinker
    * gcr.io/k8s-prow/deck
    * gcr.io/k8s-prow/horologium
    * gcr.io/k8s-prow/tide
5. Copy commit id into comment on top of the file to keep track of the current release used for the deployments (due to prow not having any releases)
6. Check with a diff tool for other meaningful changes that need to be taken into account and copied over.
7. Open [config.yaml (local)](../../prow/config.yaml) in current project as well as on the commit [config.yaml (remote)](https://github.com/kubernetes/test-infra/blob/2c8e0dbb96b4c1a86d42275dfbed5474a6d05def/prow/config.yaml)
8. Copy over new labels for the following containers:
    * gcr.io/k8s-prow/clonerefs
    * gcr.io/k8s-prow/initupload
    * gcr.io/k8s-prow/entrypoint
    * gcr.io/k8s-prow/sidecar
9. Check with a diff tool for other meaningful changes that need to be taken into account and copied over.
10. Update prow deployments
    ```bash
    kubectl apply -f prow/cluster/starter.yaml
    ```
11. Update prow config. Use the `update-config.sh {file_path}` script to apply Prow configuration on a cluster.
   ```
   ./update-config.sh ../prow/config.yaml
   ```
12. Check that everything is working as intended. E.g. that `kubectl get pods` doesn't show errors on your updated test cluster and the dashboard is still reachable.
11. Create a pull request.

## Troubleshooting

In case something goes wrong with the upgrade and pods are not starting anymore, check out the old commit before the update happened and change the deployments back to what they were via:
```bash
kubectl apply -f prow/cluster/starter.yaml
```
2. Use the following command to bring the previous config back:
```
./update-config.sh ../prow/config.yaml
```

Monitor to see the pods coming up via `kubectl get pods`
