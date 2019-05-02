#!/bin/bash

set -o errexit
set -o pipefail

source "${CURRENT_PATH}/scripts/library.sh"


shout "Lets Start!!"

shout "Create "
gcloud container clusters get-credentials $LOADGEN_CLUSTER_NAME --zone $LOADGEN_COMPUTE_ZONE --project $CLOUDSDK_CORE_PROJECT

# Create Kyma Cluster
source "${CURRENT_PATH}/scripts/cluster.sh" "--action" "create" "--cluster-grade" "production"

# Get test scripts


# Run K6 scripts
shout "Running K6 Scripts"

if [[ "${RUNMODE}" == "all" ]]; then
  shout "Running the complete test suite"
  source "k6-runner.sh" "all"
  set -o errexit
elif [[ "${RUNMODE}" == "" && "${SCRIPTPATH}" != "" ]]; then
  shout "Running following Script: $SCRIPTPATH"
  source "k6-runner.sh" "${SCRIPTPATH}"
  set -o errexit
fi


shout "Finished all k6 tests!!"

shout "Deleting the deployed kyma cluster!!"

source "${CURRENT_PATH}/scripts/cluster.sh" "--action" "delete" "--cluster-grade" "production"

