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
# Arguments:
# --excluded-clusters -  regexp of clusters that won't get removed

readonly SECONDS_PER_HOUR=3600

set -e

#Exported variables
export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"

# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/gardener/gardener.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gardener/gardener.sh"

requiredVars=(
    KYMA_PROJECT_DIR
    GARDENER_KYMA_PROW_KUBECONFIG
    GARDENER_KYMA_PROW_PROJECT_NAME
)

utils::check_required_vars "${requiredVars[@]}"

EXCLUDED_CLUSTERS_REGEX=""
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --excluded-clusters)
            EXCLUDED_CLUSTERS_REGEX="$2"
            shift
            shift
            ;;
        *)
            echo "Unknown flag ${1}"
            exit 1
            ;;
    esac
done

echo "--------------------------------------------------------------------------------"
echo "Removing Gardener clusters allocated by failed/terminated integration jobs...  "
echo "--------------------------------------------------------------------------------"

# get all cluster names in project
# shellcheck disable=SC2016
CLUSTERS=$(kubectl --kubeconfig "${GARDENER_KYMA_PROW_KUBECONFIG}" -n garden-"${GARDENER_KYMA_PROW_PROJECT_NAME}" get shoots -o go-template='{{range $i, $c :=.items}}{{if $i}},{{end}}{{$c.metadata.name}}{{end}}')
IFS="," read -r -a CLUSTERS <<< "$CLUSTERS" # convert comma separated clusters string to array

# cleanup all clusters
for CLUSTER in "${CLUSTERS[@]}"
do
    if [[ ! "$CLUSTER" =~ ${EXCLUDED_CLUSTERS_REGEX} ]]; then
        # check cluster age
        # shellcheck disable=SC2016
        CREATION_TIME="$(kubectl --kubeconfig "${GARDENER_KYMA_PROW_KUBECONFIG}" -n garden-"${GARDENER_KYMA_PROW_PROJECT_NAME}" get shoots "$CLUSTER" -o go-template='{{.metadata.creationTimestamp}}')"

        # convert to timestamp for age calculation
        CREATION_TS="$(date -d "${CREATION_TIME}" +%s)" # On macOS use: CREATION_TS=$(date -jf "%Y-%m-%dT%H:%M:%SZ" ${CREATION_TIME} +%s)
        NOW_TS="$(date +%s)"
        HOURS_OLD=$(( (NOW_TS - CREATION_TS) / SECONDS_PER_HOUR ))


        # clusters older than 24h get deleted
        # it matches clusters with day-of-week appended to the name, example: np1kyma
        if [[ ${HOURS_OLD} -ge 24 && "$CLUSTER" =~ np?[0-9].* ]]; then
            log::info "Deprovision cluster: \"${CLUSTER}\" (${HOURS_OLD}h old)"
            utils::deprovision_gardener_cluster "${GARDENER_KYMA_PROW_PROJECT_NAME}" "${CLUSTER}" "${GARDENER_KYMA_PROW_KUBECONFIG}"
        elif [[ ${HOURS_OLD} -ge 4 && ! "$CLUSTER" =~ np?[0-9].* ]]; then
            # clusters older than 4h get deleted
            log::info "Deprovision cluster: \"${CLUSTER}\" (${HOURS_OLD}h old)"
            gardener::deprovision_cluster \
                -p "${GARDENER_KYMA_PROW_PROJECT_NAME}" \
                -c "${CLUSTER}" \
                -f "${GARDENER_KYMA_PROW_KUBECONFIG}"
        fi
    else
        echo "level=warning msg=\"Cluster is excluded, deletion will be skipped. Name: \"${CLUSTER}\""
    fi
done
