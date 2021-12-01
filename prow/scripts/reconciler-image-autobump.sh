#!/usr/bin/env bash

#Description: Kyma CLI Integration plan on Gardener. This scripts implements a pipeline that consists of many steps. The purpose is to install and test Reconciler end-to-end upgrade flow on a real Gardener cluster.
#
#Expected common vars:
# - JOB_TYPE - set up by prow (presubmit, postsubmit, periodic)
# - KYMA_PROJECT_DIR - directory path with Kyma sources to use for installation
# - GARDENER_REGION - Gardener compute region
# - GARDENER_ZONES - Gardener compute zones inside the region
# - GARDENER_CLUSTER_VERSION - Version of the Kubernetes cluster
# - GARDENER_KYMA_PROW_KUBECONFIG - Kubeconfig of the Gardener service account
# - GARDENER_KYMA_PROW_PROJECT_NAME - Name of the gardener project where the cluster will be integrated.
# - GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME - Name of the secret configured in the gardener project to access the cloud provider
# - BOT_GITHUB_TOKEN: Bot github token used for API queries
# - MACHINE_TYPE - (optional) machine type
#
#Please look in each provider script for provider specific requirements

## ---------------------------------------------------------------------------------------
## Configurations and Variables
## ---------------------------------------------------------------------------------------

# exit on error, and raise error when variable is not set when used
set -e

ENABLE_TEST_LOG_COLLECTOR=false

# Exported variables
export TEST_INFRA_SOURCES_DIR="/home/prow/go/src/github.com/kyma-project/test-infra"
export CONTROL_PLANE_DIR="/home/prow/go/src/github.com/kyma-project/control-plane"
export RECONCILER_DIR="/home/prow/go/src/github.com/kyma-incubator/reconciler"

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"

# All provides require these values, each of them may check for additional variables
requiredVars=(
    RECONCILER_DIR
    CONTROL_PLANE_DIR
    GITHUB_LOGIN
    GITHUB_TOKEN_FILE
    GIT_EMAIL
    GIT_NAME
)

utils::check_required_vars "${requiredVars[@]}"

# Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"
export ERROR_LOGGING_GUARD

export IMAGE_BUMP_BRANCH="reconciler_image_autobump"
export KCP_VALUES_YAML_PATH="${CONTROL_PLANE_DIR}/resources/kcp/values.yaml"

## ---------------------------------------------------------------------------------------
## Prow job execution steps
## ---------------------------------------------------------------------------------------

log::info "Configuring git and gh"
git config --global user.email "${GIT_EMAIL}"
git config --global user.name "${GIT_NAME}"
git config --global credential.https://github.com.username "${GITHUB_LOGIN}"
git config --global credential.helper store

GIT_TOKEN_VALUE="$(cat ${GITHUB_TOKEN_FILE})"
echo "https://${GITHUB_LOGIN}:${GITHUB_TOKEN_FILE}@github.com" > ~/.git-credentials

gh config set -h github.com git_protocol https
gh config set prompt disabled
gh auth login --hostname github.com --with-token < "${GITHUB_TOKEN_FILE}"
gh auth status

log::info "Fetching latest reconciler tag"
# Checks required vars and initializes gcloud/docker if necessary
cd "${RECONCILER_DIR}"
export RECONCILER_IMAGE_TAG=$(git rev-parse --short HEAD)
echo "Bumping reconciler image to ${RECONCILER_IMAGE_TAG}"

log::info "Checking out branch (${IMAGE_BUMP_BRANCH}) in control-plane"
cd "${CONTROL_PLANE_DIR}"
git checkout -B "${IMAGE_BUMP_BRANCH}"

log::info "Updating values.yaml of KCP in control-plane"
yq e -i '.global.images.mothership_reconciler.tag = "'"${RECONCILER_IMAGE_TAG}"'"' ${KCP_VALUES_YAML_PATH}
yq e -i '.global.images.component_reconciler.tag = "'"${RECONCILER_IMAGE_TAG}"'"' ${KCP_VALUES_YAML_PATH}

log::info "Pushing changes to branch: ${IMAGE_BUMP_BRANCH}"
git commit -a -m "Bumped reconciler images to ${RECONCILER_IMAGE_TAG}"
git push --set-upstream origin ${IMAGE_BUMP_BRANCH}

log::info "Checking if any open PR exists for branch: ${IMAGE_BUMP_BRANCH}"
PR_STATUS=$(gh pr status --json state | jq -r '.currentBranch."state"')
if [[ ! ${PR_STATUS} = "OPEN" ]]; then
    log::info "Creating new PR for branch: ${IMAGE_BUMP_BRANCH}"
    gh pr create --base main --title "Reconciler image autobump [Kyma-bot]" --body "Bumped reconciler images." --label "area/reconciler"
else
    log::info "Pull Request (PR) already exists for branch: ${IMAGE_BUMP_BRANCH}"
fi

PR_URL=$(gh pr status --json url | jq -r '.currentBranch."url"')
log::info "Pull Request (PR): ${PR_URL}"

log::info "Cleaning up..."
gh auth logout --hostname github.com
rm ~/.git-credentials

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
