#!/usr/bin/env bash

#Description: Gardener cleanup job. Deletes all orphaned clusters allocated by integration jobs that coud not be successfully deleted.
# Deletes all clusters that are more than 4 hours old.
#
#
#Expected vars:
#
# - KYMA_PROJECT_DIR - directory path with Kyma sources to use for importing prow scripts
# - GARDENER_KYMA_PROW_KUBECONFIG - Kubeconfig of the Gardener service account of the project
# - GARDENER_KYMA_PROW_PROJECT_NAME Name of the gardener project where the clusters will be cleaned up.

readonly SECONDS_PER_HOUR=3600

set -e

discoverUnsetVar=false

VARIABLES=(
    KYMA_PROJECT_DIR
    GARDENER_KYMA_PROW_KUBECONFIG
    GARDENER_KYMA_PROW_PROJECT_NAME
)

for var in "${VARIABLES[@]}"; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

#Exported variables
export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"
export GARDENER_PROJECT_NAME=${GARDENER_KYMA_PROW_PROJECT_NAME}
export GARDENER_CREDENTIALS=${GARDENER_KYMA_PROW_KUBECONFIG}

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"

echo "--------------------------------------------------------------------------------"
echo "Removing Gardener clusters allocated by failed/terminated integration jobs...  "
echo "--------------------------------------------------------------------------------"

# get all cluster names in project
# shellcheck disable=SC2016
CLUSTERS=$(kubectl --kubeconfig "${GARDENER_KYMA_PROW_KUBECONFIG}" -n garden-"${GARDENER_KYMA_PROW_PROJECT_NAME}" get shoots -o go-template='{{range $i, $c :=.items}}{{if $i}},{{end}}{{$c.metadata.name}}{{end}}')
CLUSTERS=(${CLUSTERS//,/ }) # convert comma separated clusters string to array

# cleanup all clusters
for CLUSTER in "${CLUSTERS[@]}"
do
    # check cluster age
    # shellcheck disable=SC2016
    CREATION_TIME="$(kubectl --kubeconfig "${GARDENER_KYMA_PROW_KUBECONFIG}" -n garden-"${GARDENER_KYMA_PROW_PROJECT_NAME}" get shoots "$CLUSTER" -o go-template='{{.metadata.creationTimestamp}}')"

    # convert to timestamp for age calculation
    CREATION_TS="$(date -d "${CREATION_TIME}" +%s)" # On macOS use: CREATION_TS=$(date -jf "%Y-%m-%dT%H:%M:%SZ" ${CREATION_TIME} +%s)
    NOW_TS="$(date +%s)"
    HOURS_OLD=$(( (NOW_TS - CREATION_TS) / SECONDS_PER_HOUR ))
    
    # clusters older than 4h get deleted
    if [[ ${HOURS_OLD} -ge 4 ]]; then
        shout "Deprovision cluster: \"${CLUSTER}\" (${HOURS_OLD}h old)"
        date
        # Export envvars for the script
        export GARDENER_CLUSTER_NAME=${CLUSTER}
        "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/deprovision-gardener-cluster.sh
    fi
done