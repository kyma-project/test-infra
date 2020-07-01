# Prow Workload Clusters

It's a description of prow workload clusters. These are clusters where prow is scheduling pods which execute prowjob logic. All workload clusters are aggregated under kyma-prow GCP project. We use two workload clusters for trusted and untrusted prowjobs.

## Clusters design

Clusters have enabled autoupgrade and follow stable channel.
Clusters have enabled autoscaling for all node pools.
Clusters are private with unrestricted access to k8s API from Internet domain.
Clusters use CloudNAT and Private Google Access.
Clusters are regional.
Clusters use separate subnets for nodes, pods and services.

```gcloud container clusters list
   NAME                          LOCATION        MASTER_VERSION  MASTER_IP       MACHINE_TYPE   NODE_VERSION    NUM_NODES  STATUS
   trusted-workload-kyma-prow    europe-west3    1.14.10-gke.36  _____________   n1-standard-4  1.14.10-gke.36  3          RUNNING
   untrusted-workload-kyma-prow  europe-west3    1.14.10-gke.36  _____________   n1-standard-4  1.14.10-gke.36  2          RUNNING
```

## Infrastructure design

Cluster are located in separate networks for trusted and untrusted components. Each networks provide three subnets for cluster nodes, pods and services.
There is no peering between networks, thus clusters are isolated on network level.
Each cluster has dedicated Cloud Router with CloudNAT and external IP. This provide outgoing connectivity for clusters and fixed external IP from which all traffic is seen.

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
## Prow design

Prow is accessing workload clusters using x509 client certificates with cluster admin role.
Certificates are combined in to kubeconfig file and stored as a secret on prow cluster.
Jobs use context names to indicate target workload cluster to run.

```
k config get-contexts --kubeconfig workload-clusters-kubeconfig.yaml
CURRENT   NAME                 CLUSTER              AUTHINFO             NAMESPACE
*         default              default              default
          trusted-workload     trusted-workload     trusted-workload
          untrusted-workload   untrusted-workload   untrusted-workload
```
For details about building kubeconfig file and providing it to prow see [upstream documentation](https://github.com/kubernetes/test-infra/tree/master/gencred).
