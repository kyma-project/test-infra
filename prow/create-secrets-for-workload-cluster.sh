#!/bin/bash
# Source development/pre-install-prow.sh

set -o errexit

if [ -z "$PROJECT" ]; then
      echo "\$PROJECT is empty"
      exit 1
fi

if [ -z "$ZONE" ]; then
      echo "\$ZONE is empty"
      exit 1
fi

if [ -z "$WORKLOAD_CLUSTER_NAME" ]; then
      echo "\$WORKLOAD_CLUSTER_NAME is empty"
      exit 1
fi

if [ -z "$GOPATH" ]; then
      echo "\$GOPATH is not set"
      exit 1
fi

### Cloning k8s.io/test-infra is a prerequisite here
git clone "git@github.com:kubernetes/test-infra.git" "${GOPATH}/src/k8s.io/test-infra" || cd "${GOPATH}/src/k8s.io/test-infra/prow"

### Resetting to the compatible k8s.io/test-infra
git reset b9a576b397892c55487e495721d23b3a52ac9472 --hard

### Reference: https://github.com/kubernetes/test-infra/tree/master/prow/cmd/mkbuild-cluster#usage
go run "${GOPATH}/src/k8s.io/test-infra/prow/cmd/mkbuild-cluster/main.go" \
  --project="${PROJECT}" --zone="${GCLOUD_COMPUTE_ZONE}" \
  --cluster="${WORKLOAD_CLUSTER_NAME}" --alias default \
  --change-context --print-entry | tee cluster.yaml
kubectl create secret generic workload-cluster --from-file=cluster.yaml
