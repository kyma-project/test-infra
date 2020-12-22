#!/usr/bin/env bash

set -o errexit
set -o pipefail  #Fail a pipe if any sub-command fails.
########################################################################################################################
#
#Provision A New Azure EventHub Namespace in the current Azure Subscription
#
#Each Azure EventHubs Namespace can contain a maximum of 10 EventHubs (Knative Channels / Kakfa Topics) which equates
#to unique combinations of a Event Source / Event Type / Event Version. Because there is an associated cost with
#empty or unused EventHub Namespaces, we only want to provision the minimum number required.
#
#It is expected that prior to running this script the Azure subscription needs to have sufficient permissions
#to be able to perform the necessary tasks. Finally the environment should be setup with "az" and "jq"
#on their $PATH.
#
########################################################################################################################

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/azure.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/azure.sh"
# shellcheck source=prow/scripts/lib/azure.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"

#
#Global Variables
#

VARIABLES=(
  RS_GROUP
  REGION
  AZURE_SUBSCRIPTION_ID
  AZURE_CREDENTIALS_FILE
  RS_GROUP
  EVENTHUB_NAMESPACE_NAME
  EVENTHUB_SECRET_OVERRIDE_FILE
)

utils::checkRequiredVars "${VARIABLES[@]}"

EVENTHUB_NAMESPACE_MIN_THROUGHPUT_UNITS=2 #Must be greater than zero and less than maximum value!
EVENTHUB_NAMESPACE_MAX_THROUGHPUT_UNITS=4 #Must be greater than minimum value and less than 20!
EVENTHUB_NAMESPACE_LOCATION=""
EVENTHUB_NAMESPACE_SHARED_ACCESS_KEY="RootManageSharedAccessKey"

K8S_SECRET_NAME="${EVENTHUB_NAMESPACE_NAME}-overrides"
K8S_SECRET_NAMESPACE="kyma-installer"
K8S_SECRET_USERNAME="\$ConnectionString"
K8S_SECRET_PASSWORD=""

K8S_SECRET_BROKER_HOSTNAME=""
K8S_SECRET_BROKER_PORT="9093"

#
#Utility Functions To Make The Actual Cmd Line Calls
#

createGroup() {
  log::info "Create Azure group"

  az group create \
    --name "${RS_GROUP}" \
    --location "${REGION}"

  #Wait until resource group will be visible in azure.
  counter=0
  until [[ $(az group exists --name "${RS_GROUP}" -o json) == true ]]; do
    sleep 15
    counter=$(( counter + 1 ))
    if (( counter == 5 )); then
      log::error "Azure resource group ${RS_GROUP} still not present after one minute wait. Exiting..."
      exit 1
    fi
  done
}

#Create the Azure EventHubs Namespace based on global configuration
cmdCreateEventHubNamespace() {
  az eventhubs namespace create \
    --name "${EVENTHUB_NAMESPACE_NAME}" \
    --resource-group "${RS_GROUP}" \
    --location "${EVENTHUB_NAMESPACE_LOCATION}" \
    --sku standard \
    --enable-kafka true \
    --enable-auto-inflate true \
    --capacity "${EVENTHUB_NAMESPACE_MIN_THROUGHPUT_UNITS}" \
    --maximum-throughput-units "${EVENTHUB_NAMESPACE_MAX_THROUGHPUT_UNITS}" \
    --output none
}

#Get the Azure EventHub Namespace authorization keys' PrimaryConnectionString
cmdNamespacePrimaryConnectionString() {
  az eventhubs namespace authorization-rule keys list \
    -o json \
    --resource-group "${RS_GROUP}" \
    --namespace-name "${EVENTHUB_NAMESPACE_NAME}" \
    --name "${EVENTHUB_NAMESPACE_SHARED_ACCESS_KEY}" | \
    jq -r ".primaryConnectionString"
}

#Enable this while debugging to confirm the user's desire to provision
#A new EventHub Namespace for their current Azure context
confirmConfiguration() {

  #Log the configuration summary
  log::info "The following configuration will be used to provision the new EventHub Namespace - review for correctness before continuing!"
  echo " > Azure Resource Group: ${RS_GROUP}"
  echo " > New EventHub Namespace name: ${EVENTHUB_NAMESPACE_NAME}"
  echo " > New EventHub Namespace location: ${EVENTHUB_NAMESPACE_LOCATION}"
  echo " > New EventHub Namespace throughput min: ${EVENTHUB_NAMESPACE_MIN_THROUGHPUT_UNITS}"
  echo " > New EventHub Namespace throughput max: ${EVENTHUB_NAMESPACE_MAX_THROUGHPUT_UNITS}"
  echo " > Kubernetes Secret name: ${EVENTHUB_NAMESPACE_NAME}"
  echo " > Kubernetes Secret Namespace: ${K8S_SECRET_NAMESPACE}"
}

#Create the EventHub Namespace based on user's current Azure Subscription
createEventHubNamespace() {

  #Execute the Azure EventHubs Namespace creation command & handle the results
  log::info "Creating New EventHubs Namespace... (takes several minutes - be patient :)"
  if [[ $(cmdCreateEventHubNamespace) -eq 0 ]]; then
    log::info "Successfully Created New EventHub Namespace!"
  else
    log::error "Failed To Create New EventHub Namespace - Exiting!"
    exit 1
  fi
}

#Load the EventHub Namespace's authorization key information into global variables
loadAuthorizationKey() {

  #Get the new EventHub Namespace's PrimaryConnectionString
  log::info "Loading the new EventHub Namespace's authorization key..."
  local primaryConnectionString=""
  primaryConnectionString=$(cmdNamespacePrimaryConnectionString)

  #Populate the Kubernetes Secret Broker / Password Values
  K8S_SECRET_BROKER_HOSTNAME=$(echo "${primaryConnectionString}" | sed -e "s/^Endpoint=.*sb:\/\/\(.*\)\/;.*$/\1/")
  K8S_SECRET_PASSWORD=${primaryConnectionString}
}

#Creates The EventHub Namespace Secret override file
createK8SSecretFile() {

  log::info "Creating a Kubernetes Secret override file for the New EventHub Namespace..."

kafkaSecret=$(cat << EOF
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: "${K8S_SECRET_NAME}"
  namespace: "${K8S_SECRET_NAMESPACE}"
  labels:
    knativekafka.kyma-project.io/kafka-secret: "true"
    installer: overrides
    component: knative-eventing-kafka
    kyma-project.io/installation: ""
stringData:
  kafka.brokers.hostname: "${K8S_SECRET_BROKER_HOSTNAME}"
  kafka.brokers.port: "${K8S_SECRET_BROKER_PORT}"
  kafka.namespace: "${EVENTHUB_NAMESPACE_NAME}"
  kafka.password: "${K8S_SECRET_PASSWORD}"
  kafka.username: "${K8S_SECRET_USERNAME}"
  kafka.secretName: "knative-kafka"
  environment.kafkaProvider: "azure"
EOF
)
  echo "${kafkaSecret}" > "${EVENTHUB_SECRET_OVERRIDE_FILE}"
}

#
#Main Script Execution
#

# Login to azure using credentials file
az::login "${AZURE_CREDENTIALS_FILE}"
az::set_subscription "${AZURE_SUBSCRIPTION_ID}"

#Confirm the configuration
confirmConfiguration

#Create the New Azure Resource Group
createGroup

#Create the New Azure EventHubs Namespace
createEventHubNamespace

#Lookup the New Azure EventHub Namespace's authorization key
loadAuthorizationKey

#Create K8S Secret override file for EventHubs Namespace
createK8SSecretFile
