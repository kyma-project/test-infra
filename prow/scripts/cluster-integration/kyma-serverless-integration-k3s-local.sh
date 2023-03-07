#!/usr/bin/env bash

set -o errexit
set -o pipefail  # Fail a pipe if any sub-command fails.

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# shellcheck source=prow/scripts/lib/serverless-shared-k3s.sh
source "${SCRIPT_DIR}/../lib/serverless-shared-k3s.sh"

date

export DOMAIN=${KYMA_DOMAIN:-local.kyma.dev}
if [[ -z $REGISTRY_VALUES ]]; then
  export REGISTRY_VALUES="dockerRegistry.enableInternal=false,dockerRegistry.serverAddress=registry.localhost:5000,dockerRegistry.registryAddress=registry.localhost:5000"
fi

export KYMA_SOURCES_DIR="./kyma"
export SERVERLESS_OVERRIDES_DIR="./overrides"

export INTEGRATION_SUITE=${1:-serverless-integration}

echo "--> Installing kyma-cli"
install::kyma_cli

echo "--> Provisioning k3d cluster for Kyma"
kyma provision k3d --ci

echo "--> Deploying Serverless"
# The python38 function requires 40M+ of memory to work. Mostly used by kubeless. I need to overrride the defaultPreset to M to avoid OOMkill.

if [[ ${INTEGRATION_SUITE} == "git-auth-integration" ]]; then
  echo "--> Deploying Serverless from Kyma main"
  kyma deploy -p evaluation --ci \
    --component cluster-essentials \
    --component serverless \
    --value "$REGISTRY_VALUES" \
    --value global.ingress.domainName="$DOMAIN" \
    --value "serverless.webhook.values.function.resources.defaultPreset=M" \
    --value "serverless.webhook.values.featureFlags.java17AlphaEnabled=true" \
    -s main
else
  echo "--> Deploying Serverless from $KYMA_SOURCES_DIR"
  kyma deploy -p evaluation --ci \
    --component cluster-essentials \
    --component serverless \
    --value "$REGISTRY_VALUES" \
    --value global.ingress.domainName="$DOMAIN" \
    --value "serverless.webhook.values.function.resources.defaultPreset=M" \
    --value "serverless.webhook.values.featureFlags.java17AlphaEnabled=true" \
    -s local -w $KYMA_SOURCES_DIR
fi

echo "##############################################################################"
# shellcheck disable=SC2004
echo "# Serverless installed in $(( $SECONDS/60 )) min $(( $SECONDS % 60 )) sec"
echo "##############################################################################"

# TODO: I can consider to remove this and use loop for ready webhook and operator
# I know it's bad practice and kinda smelly to do this, but we have two nasty dataraces that might happen, and simple sleep solves them both:
# webhook might not be ready in time (but somehow it still accepts the function, we have an issue for that)
# runtime configmaps might now have been copied to that namespace, but it should be handled by https://github.com/kyma-project/kyma/pull/10026
########
sleep 60
########

SERVERLESS_CHART_DIR="${KYMA_SOURCES_DIR}/resources/serverless"
job_name="k3s-serverless-test"

if [[ ${INTEGRATION_SUITE} == "git-auth-integration" ]]; then
  echo "--> Fetching Serverless k3s-tests"
 
  git clone https://github.com/kyma-project/kyma "${KYMA_SOURCES_DIR}"
  job_name="k3s-serverless-nightly-test"
fi

VALUES="-f ${SERVERLESS_CHART_DIR}/values.yaml"
# check for test overrides.
if [[ -e "${SERVERLESS_OVERRIDES_DIR}/integration-overrides.yaml" ]]; then
  VALUES+=" -f ${SERVERLESS_OVERRIDES_DIR}/integration-overrides.yaml"
fi

#TODO: This is the part of helm running the test, we need to remove it
#shellcheck disable=SC2086
make $(INTEGRATION_SUITE)
job_status=$?

collect_results "${job_name}" "default"

echo "Exit code ${job_status}"

exit $job_status
