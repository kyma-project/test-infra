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

if [[ ${INTEGRATION_SUITE} == "git-auth-integration" ]]; then
  echo "--> Cloning Serverless integration tests from kyma:main"
  git clone https://github.com/kyma-project/kyma "${KYMA_SOURCES_DIR}"
fi

# check for test secrets.
if [[ -e "${SERVERLESS_OVERRIDES_DIR}/git-auth.env" ]]; then
  # shellcheck source=/dev/null
  source "${SERVERLESS_OVERRIDES_DIR}/git-auth.env"
fi

export APP_TEST_CLEANUP="onSuccessOnly"
#https://github.com/kyma-project/test-infra/issues/6513
export PATH=${PATH}:/usr/local/go/bin
set +o errexit
(cd ${KYMA_SOURCES_DIR}/tests/function-controller && make "${INTEGRATION_SUITE}")
job_status=$?
set -o errexit

collect_results

echo "Exit code ${job_status}"

exit $job_status
