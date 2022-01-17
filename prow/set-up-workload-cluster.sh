#!/bin/bash
# Source development/prow/set-up-workload-cluster.sh
set -o errexit

### 
readonly CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly KUBECONFIG=${KUBECONFIG:-"${HOME}/.kube/config"}
readonly PROW_WORKLOAD_CLUSTER_DIR="$( cd "${CURRENT_DIR}/workload-cluster" && pwd )"


if [ -z "$CLUSTER_NAME" ]; then
      echo "\$CLUSTER_NAME is empty"
      exit 1
fi

# requried by External Secrets Syncer to access Secret Manager
if [ -z "$GOOGLE_APPLICATION_CREDENTIALS" ]; then
      echo "\$GOOGLE_APPLICATION_CREDENTIALS is empty"
      exit 1
fi

# Set up ClusterRoleBinding for User: Client which plank needs to operate on this cluster
kubectl apply -f "${PROW_WORKLOAD_CLUSTER_DIR}/00-clusterrolebinding.yaml"

# Install PodDisruptionBudgets
kubectl apply -f "${PROW_WORKLOAD_CLUSTER_DIR}/02-kube-system_poddisruptionbudgets.yaml"

# Overwrite kube-dns-autoscaler config map
cat <<EOF | kubectl replace -f -
apiVersion: v1
data:
  linear: '{"coresPerReplica":256,"nodesPerReplica":8,"preventSinglePointFailure":true}'
kind: ConfigMap
metadata:
  name: kube-dns-autoscaler
  namespace: kube-system
EOF

# Configure stub-domains to speed up DNS propagation
kubectl -n kube-system patch cm kube-dns --type merge --patch \
  "$(cat "${PROW_WORKLOAD_CLUSTER_DIR}"/03-kube-dns-stub-domains-patch.yaml)"


# Create secrets
# namespace
kubectl create namespace external-secrets
# Service account
kubectl create secret generic sa-secret-manager-"$CLUSTER_NAME" --namespace "external-secrets" --from-file=service-account.json="$GOOGLE_APPLICATION_CREDENTIALS"
# install helm chart
helm repo add external-secrets https://external-secrets.github.io/kubernetes-external-secrets/
helm install 8.2.1 external-secrets/kubernetes-external-secrets
helm install -f "${CLUSTER_DIR}/resources/external-secrets/values_${CLUSTER_NAME}.yaml" -n external-secrets kubernetes-external-secrets external-secrets/kubernetes-external-secrets
# apply external secrets used by all workload clusters
kubectl apply -f "${CLUSTER_DIR}/resources/external-secrets/external_secrets_workloads.yaml"
# apply additional external secrets if they exist
if [[ -f "${CLUSTER_DIR}/resources/external-secrets/external_secrets_${CLUSTER_NAME}.yaml" ]]; then
      kubectl apply -f "${CLUSTER_DIR}/resources/external-secrets/external_secrets_${CLUSTER_NAME}.yaml"
fi
