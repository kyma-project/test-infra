#!/usr/bin/env bash

#Description: Kyma CLI Integration plan on Gardener. This scripts implements a pipeline that consists of many steps. The purpose is to install and test Kyma using the CLI on a real Gardener cluster.
#
#Expected common vars:
# - JOB_TYPE - set up by prow (presubmit, postsubmit, periodic)
# - KYMA_PROJECT_DIR - directory path with Kyma sources to use for installation
# - GARDENER_REGION - Gardener compute region
# - GARDENER_ZONES - Gardener compute zones inside the region
# - GARDENER_CLUSTER_VERSION - Version of the Kubernetes cluster
# - GARDENER_KYMA_PROW_KUBECONFIG - Kubeconfig of the Gardener service account
# - GARDENER_KYMA_PROW_PROJECT_NAME - Name of the gardener project where the cluster will be integrated
# - GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME - Name of the secret configured in the gardener project to access the cloud provider
# - CREDENTIALS_DIR - Directory where is the EventMesh service key is mounted
# - MACHINE_TYPE - (optional) machine type
#
#Please look in each provider script for provider specific requirements

## ---------------------------------------------------------------------------------------
## Configurations and Variables
## ---------------------------------------------------------------------------------------

# exit on error, and raise error when variable is not set when used
set -e

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"
export EVENTMESH_SECRET_FILE="${CREDENTIALS_DIR}/serviceKey" # For eventing E2E fast-integration tests

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/kyma.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"
# shellcheck source=prow/scripts/cluster-integration/helpers/eventing.sh
source "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/eventing.sh"

# All provides require these values, each of them may check for additional variables
requiredVars=(
    GARDENER_PROVIDER
    KYMA_PROJECT_DIR
    GARDENER_REGION
    GARDENER_ZONES
    GARDENER_CLUSTER_VERSION
    GARDENER_KYMA_PROW_KUBECONFIG
    GARDENER_KYMA_PROW_PROJECT_NAME
    GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME
    CREDENTIALS_DIR
    EVENTMESH_SECRET_FILE
    TEST_EVENTING_AUTH_IAS_URL
    TEST_EVENTING_AUTH_IAS_USER
    TEST_EVENTING_AUTH_IAS_PASSWORD
)

utils::check_required_vars "${requiredVars[@]}"

if [[ $GARDENER_PROVIDER == "azure" ]]; then
    # shellcheck source=prow/scripts/lib/gardener/azure.sh
    source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gardener/azure.sh"
elif [[ $GARDENER_PROVIDER == "aws" ]]; then
    # shellcheck source=prow/scripts/lib/gardener/aws.sh
    source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gardener/aws.sh"
elif [[ $GARDENER_PROVIDER == "gcp" ]]; then
    # shellcheck source=prow/scripts/lib/gardener/gcp.sh
    source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gardener/gcp.sh"
    # shellcheck source=prow/scripts/lib/gardener/gardener.sh
    source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gardener/gardener.sh"
else
    ## TODO what should I put here? Is this a backend?
    log::error "GARDENER_PROVIDER ${GARDENER_PROVIDER} is not yet supported"
    exit 1
fi

# needed variables to communicate with the CIS APIs
location=""
user=""

function cleanupJobAssets() {
    # Must be at the beginning
    EXIT_STATUS=$?

    set +e

    log::banner "Job Exit Status:: \"${EXIT_STATUS}\""

    # delete application
    curl -X DELETE "${TEST_EVENTING_AUTH_IAS_URL}${location}" --user "${user}"

    if [[ $EXIT_STATUS != "0" ]]; then
        eventing::print_troubleshooting_logs
    fi

    log::banner "Cleanup fast-integration assets"
    eventing::fast_integration_test_cleanup

    log::banner "Cleaning job assets"
    if  [[ "${CLEANUP_CLUSTER}" == "true" ]] ; then
        log::info "Deprovision cluster: \"${CLUSTER_NAME}\""
        gardener::deprovision_cluster \
            -p "${GARDENER_KYMA_PROW_PROJECT_NAME}" \
            -c "${CLUSTER_NAME}" \
            -f "${GARDENER_KYMA_PROW_KUBECONFIG}"
    fi

    set -e
    exit ${EXIT_STATUS}
}

# nice cleanup on exit, be it successful or on fail
trap cleanupJobAssets EXIT INT

#Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"
export ERROR_LOGGING_GUARD

RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c6)
readonly COMMON_NAME_PREFIX="grd"
COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
export COMMON_NAME

export CLUSTER_NAME="${COMMON_NAME}"

# set KYMA_SOURCE used by gardener::deploy_kyma
utils::generate_vars_for_build \
    -b "$BUILD_TYPE" \
    -p "$PULL_NUMBER" \
    -s "$PULL_BASE_SHA" \
    -n "$JOB_NAME"
export KYMA_SOURCE=${utils_generate_vars_for_build_return_kymaSource:?}

## ---------------------------------------------------------------------------------------
## Prow job execution steps
## ---------------------------------------------------------------------------------------

# checks required vars and initializes gcloud/docker if necessary
gardener::init

# if MACHINE_TYPE is not set then use default one
gardener::set_machine_type

kyma::install_cli

# currently only Azure generates overrides, but this may change in the future
gardener::generate_overrides

export CLEANUP_CLUSTER="true"
gardener::provision_cluster

## ---------------------------------------------------------------------------------------
## Create the Eventing webhook auth secret
## ---------------------------------------------------------------------------------------

log::banner "Create secret kyma-system/eventing-webhook-auth"

user="${TEST_EVENTING_AUTH_IAS_USER}:${TEST_EVENTING_AUTH_IAS_PASSWORD}"

# generate uuid
uuid=$(cat /proc/sys/kernel/random/uuid)

# create application
location=$(curl "${TEST_EVENTING_AUTH_IAS_URL}/Applications/v1/" \
--silent \
--include \
--header 'Content-Type: application/json' \
--user "${user}" \
--data '{
  "id": "'"${uuid}"'",
  "name": "'"${uuid}"'",
  "multiTenantApp": false,
  "schemas": [
    "urn:sap:identity:application:schemas:core:1.0",
    "urn:sap:identity:application:schemas:extension:sci:1.0:Authentication"
  ],
  "branding" : {
    "displayName": "'"${KYMA_SOURCE}"' - Kyma integration gardener Eventing"
  },
  "urn:sap:identity:application:schemas:extension:sci:1.0:Authentication": {
    "ssoType": "openIdConnect",
    "subjectNameIdentifier": "uid",
    "rememberMeExpirationTimeInMonths": 3,
    "passwordPolicy": "https://accounts.sap.com/policy/passwords/sap/web/1.1",
    "userAccess": {
      "type": "internal",
      "userAttributesForAccess": [
        {
          "userAttributeName": "firstName",
          "isRequired": false
        },
        {
          "userAttributeName": "lastName",
          "isRequired": true
        },
        {
          "userAttributeName": "mail",
          "isRequired": true
        }
      ]
    },
    "assertionAttributes": [
      {
        "assertionAttributeName": "first_name",
        "userAttributeName": "firstName"
      },
      {
        "assertionAttributeName": "last_name",
        "userAttributeName": "lastName"
      },
      {
        "assertionAttributeName": "mail",
        "userAttributeName": "mail"
      },
      {
        "assertionAttributeName": "user_uuid",
        "userAttributeName": "userUuid"
      },
      {
        "assertionAttributeName": "locale",
        "userAttributeName": "language"
      }
    ],
    "spnegoEnabled": false,
    "biometricAuthenticationEnabled": false,
    "verifyMail": true,
    "forceAuthentication": false,
    "trustAllCorporateIdentityProviders": false,
    "allowIasUsers": false,
    "riskBasedAuthentication": {
      "defaultAction": [
        "allow"
      ]
    },
    "saml2Configuration": {
      "defaultNameIdFormat": "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified",
      "signSLOMessages": true,
      "requireSignedSLOMessages": true,
      "requireSignedAuthnRequest": false,
      "signAssertions": true,
      "signAuthnResponses": false,
      "responseElementsToEncrypt": "none",
      "digestAlgorithm": "sha256",
      "proxyAuthnRequest": {
        "authenticationContext": "none"
      }
    },
    "openIdConnectConfiguration": {}
  }
}
' | grep -i ^Location: | cut -d: -f2- | sed 's/^ *\(.*\).*/\1/' | tr -d '[:space:]')

exit_status=$?
if [ ${exit_status} -ne 0 ]; then
  log::error "Error occurred while creating the application"
  exit 1
fi

# get client id
client_id=$(curl --silent "${TEST_EVENTING_AUTH_IAS_URL}${location}" \
--user "${user}" | jq -r '.["urn:sap:identity:application:schemas:extension:sci:1.0:Authentication"].clientId')

# generate client secret
client_secret=$(curl --silent "${TEST_EVENTING_AUTH_IAS_URL}${location}/apiSecrets" \
--header 'Content-Type: application/json' \
--user "${user}" \
--data '{
  "authorizationScopes": [
    "oAuth",
    "manageApp",
    "manageUsers"
  ],
  "validTo": "2030-01-01T10:00:00Z"
}' | jq -r '.secret')

# generate token and certs url
token_url="${TEST_EVENTING_AUTH_IAS_URL}/oauth2/token"
certs_url="${TEST_EVENTING_AUTH_IAS_URL}/oauth2/certs"

# create eventing webhook auth secret
kubectl apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: kyma-system
---
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  namespace: kyma-system
  name: eventing-webhook-auth
stringData:
  certs_url: "${certs_url}"
  token_url: "${token_url}"
  client_id: "${client_id}"
  client_secret: "${client_secret}"
---
EOF

## ---------------------------------------------------------------------------------------
## Deploy & Run Kyma tests
## ---------------------------------------------------------------------------------------

# deploy Kyma
log::info "Deploying Kyma ${KYMA_SOURCE}"
gardener::deploy_kyma -p "$EXECUTION_PROFILE" --source "${KYMA_SOURCE}" --value eventing.controller.eventingWebhookAuth.enabled=true

# generate pod-security-policy list in json
utils::save_psp_list "${ARTIFACTS}/kyma-psp.json"


if [[ "${HIBERNATION_ENABLED}" == "true" ]]; then
    gardener::hibernate_kyma
    # TODO make the sleep value configurable (if it makes sense)
    sleep 120
    gardener::wake_up_kyma
fi

eventing::print_subscription_crd_version

if [[ "${EXECUTION_PROFILE}" == "evaluation" ]] || [[ "${EXECUTION_PROFILE}" == "production" ]]; then
    # test the default Eventing backend which comes with Kyma
    log::banner "Execute eventing E2E fast-integration tests"
    # eventing test assets cleanup will done later in cleanup script
    eventing::test_fast_integration_eventing_prep
    eventing::fast_integration_tests
else
    gardener::test_kyma
fi

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
