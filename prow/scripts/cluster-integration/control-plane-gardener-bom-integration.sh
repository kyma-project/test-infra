#!/usr/bin/env bash

set -o errexit
set -o pipefail  # Fail a pipe if any sub-command fails.

# Required OIDC Envc
# OIDC_ISSUER_URL
# OIDC_CLIENT_ID
# OIDC_CLIENT_SECRET

# TODO: add temporary dummy config
OIDC_ISSUER_URL="https://test"
OIDC_CLIENT_ID="abcd"
OIDC_CLIENT_SECRET="efgh"

VARIABLES=(
    KYMA_PROJECT_DIR
    GARDENER_KYMA_PROW_KUBECONFIG
    GARDENER_KYMA_PROW_PROJECT_NAME
    GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME
)

discoverUnsetVar=false
for var in "${VARIABLES[@]}"; do
  if [ -z "${!var}" ] ; then
    echo "ERROR: $var is not set"
    discoverUnsetVar=true
  fi
done
if [ "${discoverUnsetVar}" = true ] ; then
  exit 1
fi

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export KCP_SOURCES_DIR="${KYMA_PROJECT_DIR}/control-plane"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"

export GARDENER_PROJECT_NAME="${GARDENER_KYMA_PROW_PROJECT_NAME}"
export GARDENER_APPLICATION_CREDENTIALS="${GARDENER_KYMA_PROW_KUBECONFIG}"
export GARDENER_AZURE_SECRET_NAME="${GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME}"


readonly RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c5)
export CLUSTER_NAME="bom-test-${RANDOM_NAME_SUFFIX}"

#!Put cleanup code in this function! Function is executed at exit from the script and on interuption.
cleanup() {
    #!!! Must be at the beginning of this function !!!
    EXIT_STATUS=$?
    #Turn off exit-on-error so that next step is executed even if previous one fails.
    set +e

    # TODO: uncomment when running tests
    # if [[ -n "${SUITE_NAME}" ]]; then
    #     testSummary
    # fi 

    if [ "${ERROR_LOGGING_GUARD}" = "true" ]; then
        shout "AN ERROR OCCURED! Take a look at preceding log entries."
        echo
    fi

    if [ -n "${CLEANUP_CLUSTER}" ]; then
        shout "Deprovision cluster: \"${CLUSTER_NAME}\""
        date
        # Export envvars for the script
        export GARDENER_CLUSTER_NAME=${CLUSTER_NAME}
        export GARDENER_PROJECT_NAME=${GARDENER_KYMA_PROW_PROJECT_NAME}
        export GARDENER_CREDENTIALS=${GARDENER_KYMA_PROW_KUBECONFIG}
        "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/deprovision-gardener-cluster.sh
    fi

    rm -rf "${TMP_DIR}"

    MSG=""
    if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
    shout "Job is finished ${MSG}"
    date
    set -e

    exit "${EXIT_STATUS}"
}


export KUBECONFIG=${GARDENER_APPLICATION_CREDENTIALS}

KCP_PROVISIONER_DIR="${KCP_SOURCES_DIR}/components/provisioner"

# Generate yamls
go run ${KCP_PROVISIONER_DIR}/cmd/template render \ # TODO oidc flags with vals
 --shoot ${CLUSTER_NAME} --project ${GARDENER_PROJECT_NAME} --secret ${GARDENER_AZURE_SECRET_NAME} \
 --gardener-domain "canary.k8s.ondemand.com" --oidc-issuer-url ${OIDC_ISSUER_URL} \
 --oidc-client-id ${OIDC_CLIENT_ID} --oidc-client-secret ${OIDC_CLIENT_SECRET}

kubectl apply -f ${KCP_PROVISIONER_DIR}/templates-rendered/shoot.yaml
kubectl apply -f ${KCP_PROVISIONER_DIR}/templates-rendered/cluster-bom.yaml

# Wait 1 min for Gardener to start provisioning before checking state
sleep 60

shoot "Waiting for Shoot operation to be in Succeeded state"

state=""
maxTries=120 # +- 40 minutes
tries=0

while [ $state != 'Succeeded' ] && [ $tries -lt $maxTries ]
do
    echo "waiting for Shoot to be provisioned..."
    state=$(kubectl get shoot ${CLUSTER_NAME} -o jsonpath='{.status.lastOperation.state}')
    tries=$((tries+1))
    sleep 20
done

shoot "Shoot provisioned"

# Wait 1 min for Gardener to start Kyma installation before checking state
sleep 60

shout "Installation triggered"
date
"${KCP_SCRIPTS_DIR}"/is-installed.sh --timeout 45m

## TODO: verify which tests can be run on such scenario 

# shout "Test Kyma"
# date
# "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/kyma-testing.sh

shout "Success"

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"