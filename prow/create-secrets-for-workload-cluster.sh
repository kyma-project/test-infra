#!/bin/bash
# Source development/pre-install-prow.sh

set -o errexit

if [ -z "$PROJECT" ]; then
      echo "\$PROJECT is empty"
      exit 1
fi

if [ -z "$GCLOUD_COMPUTE_ZONE" ]; then
      echo "\$GCLOUD_COMPUTE_ZONE is empty"
      exit 1
fi

if [ -z "$CLUSTER_NAME" ]; then
      echo "\$CLUSTER_NAME is empty"
      exit 1
fi

if [ -z "$GOPATH" ]; then
      echo "\$GOPATH is not set"
      exit 1
fi

### Cloning k8s.io/test-infra is a prerequisite here
git clone "git@github.com:kubernetes/test-infra.git" "${GOPATH}/src/k8s.io/test-infra" || cd $GOPATH/src/k8s.io/test-infra/prow

### Reference: https://github.com/kubernetes/test-infra/tree/master/prow/cmd/mkbuild-cluster#usage
bazel run //prow/cmd/mkbuild-cluster -- \
  --project="${PROJECT}" --zone="${GCLOUD_COMPUTE_ZONE}" --cluster="${CLUSTER_NAME}" --alias default --change-context --print-entry | tee cluster.yaml
kubectl create secret generic workload-cluster --from-file=cluster.yaml