# Prow Workload Clusters

This document describes workload clusters on which Prow schedules Pods to execute the logic of a given Prow job. All workload clusters are aggregated under the `kyma-prow` Google Cloud project. We use two workload clusters for trusted and untrusted Prow jobs.

## Clusters Design

Workload clusters:
- Have autoupgrade enabled and follow a stable channel.
- Have autoscaling enabled for all node pools.
- Are [private](https://cloud.google.com/kubernetes-engine/docs/concepts/private-cluster-concept), and have unrestricted access to the k8s API from the Internet domain.
- Use Cloud NAT and Private Google Access.
- Are regional.
- Use separate subnets for nodes, Pods, and services.

```gcloud container clusters list
   NAME                          LOCATION        MASTER_VERSION  MASTER_IP       MACHINE_TYPE   NODE_VERSION    NUM_NODES  STATUS
   trusted-workload-kyma-prow    europe-west3    1.14.10-gke.36  _____________   n1-standard-4  1.14.10-gke.36  3          RUNNING
   untrusted-workload-kyma-prow  europe-west3    1.14.10-gke.36  _____________   n1-standard-4  1.14.10-gke.36  2          RUNNING
```

## Infrastructure Design

Clusters are located in separate networks for trusted and untrusted components. Each network provides three subnets for cluster nodes, Pods, and services.
There is no peering between networks, thus clusters are isolated on the network level.
Each cluster has a dedicated Cloud Router with Cloud NAT and an external IP address. This provides outgoing connectivity for clusters and a fixed external IP to see all traffic originated from Pods and worker nodes.

```
gcloud compute networks list
NAME                 SUBNET_MODE  BGP_ROUTING_MODE  IPV4_RANGE  GATEWAY_IPV4
trusted-kyma-prow    CUSTOM       GLOBAL
untrusted-kyma-prow  CUSTOM       GLOBAL

gcloud compute networks subnets list
NAME                          REGION                   NETWORK              RANGE
trusted-workload-kyma-prow    europe-west3             trusted-kyma-prow    
untrusted-workload-kyma-prow  europe-west3             untrusted-kyma-prow  

gcloud compute routers list
NAME                          REGION        NETWORK
trusted-workload-kyma-prow    europe-west3  trusted-kyma-prow
untrusted-workload-kyma-prow  europe-west3  untrusted-kyma-prow

gcloud compute addresses list
NAME                          ADDRESS/RANGE   TYPE      PURPOSE  NETWORK  REGION        SUBNET  STATUS
trusted-workload-kyma-prow    _____________   EXTERNAL                    europe-west3          IN_USE
untrusted-workload-kyma-prow  _____________   EXTERNAL                    europe-west3          IN_USE
```
## Prow Design

Prow accesses workload clusters using X.509 client certificates and the **cluster-admin** role.
Certificates are combined into a kubeconfig file and stored as a secret on a Prow cluster.
Jobs use context names to indicate the target workload cluster to run.

```
k config get-contexts --kubeconfig workload-clusters-kubeconfig.yaml
CURRENT   NAME                 CLUSTER              AUTHINFO             NAMESPACE
*         default              default              default
          trusted-workload     trusted-workload     trusted-workload
          untrusted-workload   untrusted-workload   untrusted-workload
```
For details about building the kubeconfig file and providing it to Prow, see the official [k8s documentation](https://github.com/kubernetes/test-infra/tree/master/gencred).
