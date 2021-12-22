#!/usr/bin/env bash

set -e

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
KYMA_RESOURCES_DIR="/home/prow/go/src/github.com/kyma-project/kyma/installation/resources"

# shellcheck source=prow/scripts/lib/docker.sh
source "$SCRIPT_DIR/lib/docker.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
# source "$SCRIPT_DIR/lib/gcp.sh"


# gcp::authenticate \
#   -c "$SA_KYMA_ARTIFACTS_GOOGLE_APPLICATION_CREDENTIALS"

docker::authenticate "${GCR_PUSH_GOOGLE_APPLICATION_CREDENTIALS}"


echo "This tool generates component descriptor file"
/prow-tools/image-url-helper \
    --resources-directory /home/prow/go/src/github.com/kyma-project/kyma/resources/ \
    components \
    --component-version $(date +v%Y%m%d)-${PULL_PULL_SHA::8} \
    --git-commit ${PULL_PULL_SHA} \
    --git-branch ${PULL_BASE_REF} \
    --output-dir ${ARTIFACTS}/cd \
    --repo-context $(DOCKER_PUSH_REPOSITORY)
echo "Compomnent decriptor was generated succesfully finished"
