#!/bin/bash
# Source development/prow/set-up-workload-cluster.sh
set -o errexit

### 
readonly CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly KUBECONFIG=${KUBECONFIG:-"${HOME}/.kube/config"}
readonly PROW_WORKLOAD_CLUSTER_DIR="$( cd "${CURRENT_DIR}/workload-cluster" && pwd )"

if [ -z "$BUCKET_NAME" ]; then
      echo "\$BUCKET_NAME is empty"
      exit 1
fi

if [ -z "$KEYRING_NAME" ]; then
      echo "\$KEYRING_NAME is empty"
      exit 1
fi

if [ -z "$ENCRYPTION_KEY_NAME" ]; then
      echo "\$ENCRYPTION_KEY_NAME is empty"
      exit 1
fi

if [ -z "${LOCATION}" ]; then
    LOCATION="global"
fi

# requried by secretspopulator
if [ -z "$GOOGLE_APPLICATION_CREDENTIALS" ]; then
      echo "\$GOOGLE_APPLICATION_CREDENTIALS is empty"
      exit 1
fi

if [ -z "$PROJECT" ]; then
      echo "\$PROJECT is empty"
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
go run "${CURRENT_DIR}/../development/tools/cmd/secretspopulator/main.go" --project="${PROJECT}" --location "${LOCATION}" --bucket "${BUCKET_NAME}" --keyring "${KEYRING_NAME}" --key "${ENCRYPTION_KEY_NAME}" --kubeconfig "${KUBECONFIG}" --secrets-def-file="${PROW_WORKLOAD_CLUSTER_DIR}/required-secrets.yaml"

