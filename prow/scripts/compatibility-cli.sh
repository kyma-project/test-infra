#!/usr/bin/env bash

###
# This script is designed to check the compatibility of the Kyma CLI with previous Kyma versions.
# It will calculate which Kyma version should be tested and will pass it on to the 'provision-vm-cli.sh' script
#
# INPUT:
# - COMPAT_BACKTRACK: number of versions to go back in Kyma
#
# REQUIREMENTS:
# - git
###

set -o errexit

VARIABLES=(
    COMPAT_BACKTRACK
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
export KYMA_PROJECT_DIR=${KYMA_PROJECT_DIR:-"/home/prow/go/src/github.com/kyma-project"}
export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export TEST_INFRA_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts"

# shellcheck disable=SC1090
source "${TEST_INFRA_SCRIPTS}/library.sh"

echo "--------------------------------------------------------------------------------"
echo "Kyma CLI compatibility checker"
echo "--------------------------------------------------------------------------------"

# Get all non RC release tags
RELEASES=$(curl -s https://api.github.com/repos/kyma-project/kyma/releases | grep -E '(tag_name": ")(\d+\.\d+\.\d+\",?$)' | awk '{print $2}')

# Clean up spaces and quotes
RELEASES=${RELEASES//[,\"]/ }

# Split into array
RELEASES=(${RELEASES})
echo "Releases is: ${RELEASES[@]}"

# Go through releases ignoring patch versions in descending order until we skip the desired number of minor releases

# remove patch
CURRENT=$(echo "${RELEASES[0]}" | awk -F'.' '{print $1"."$2}')
for r in "${RELEASES[@]}"; do
    # remove patch from candidate
    WANT=$(echo "${r}" | awk -F'.' '{print $1"."$2}')

    if [[ "$WANT" != "$CURRENT" ]]; then
        # check if we need to backtrack more
        if [[ $COMPAT_BACKTRACK == 1 ]]; then
            # Found the target release
            TARGET=$r
            break
        else
            # Still need to backtrack further
            COMPAT_BACKTRACK=$((COMPAT_BACKTRACK - 1))
            CURRENT=$(echo "${r}" | awk -F'.' '{print $1"."$2}')
        fi
    fi
done

shout "Checking Kyma CLI compatibility with Kyma ${TARGET}"
date

# Call CLI integration script with the target release
"${TEST_INFRA_SCRIPTS}"/provision-vm-cli.sh --kyma-version "${TARGET}"