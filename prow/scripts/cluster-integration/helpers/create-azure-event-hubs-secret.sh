#!/usr/bin/env bash

set -o errexit
set -o pipefail  # Fail a pipe if any sub-command fails.

#!/bin/bash
########################################################################################################################
#
# Provision A New Azure EventHub Namespace In The Current Azure Subscription
#
# Each Azure EventHubs Namespace can contain a maximum of 10 EventHubs (Knative Channels / Kakfa Topics) which equates
# to unique combinations of a Event Source / Event Type / Event Version.  Because there is an associated cost with
# empty or unused EventHub Namespaces, we only want to provision the minimum number required.
#
# It is expected that prior to running this script the Azure subscription needs to have sufficient permissions
# to be able to perform the necessary tasks.  Finally the environment should be setup with "az", "kubectl" and "jq"
# on their $PATH.
#
########################################################################################################################

set -e

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"

#
# Global Variables
#

VARIABLES=(
	RS_GROUP
	REGION
	AZURE_SUBSCRIPTION_ID
	AZURE_SUBSCRIPTION_APP_ID
	AZURE_SUBSCRIPTION_SECRET
	AZURE_SUBSCRIPTION_TENANT
	RS_GROUP
	EVENTHUB_NAMESPACE_NAME
)

for var in "${VARIABLES[@]}"; do
	if [ -z "${!var}" ] ; then
		shout "ERROR: $var is not set"
		discoverUnsetVar=true
	fi
done

if [ "${discoverUnsetVar}" = true ] ; then
	exit 1
fi

EVENTHUB_NAMESPACE_MIN_THROUGHPUT_UNITS=2 # Must be greater than zero and less than maximum value!
EVENTHUB_NAMESPACE_MAX_THROUGHPUT_UNITS=4 # Must be greater than minimum value and less than 20!
EVENTHUB_NAMESPACE_LOCATION=""
EVENTHUB_NAMESPACE_SHARED_ACCESS_KEY="RootManageSharedAccessKey"
K8S_SECRET_NAME="${EVENTHUB_NAMESPACE_NAME}-overrides"
K8S_SECRET_NAMESPACE="kyma-installer"
K8S_SECRET_USERNAME="\$ConnectionString"
K8S_SECRET_BROKER=""
K8S_SECRET_PASSWORD=""

#
# Utility Functions To Make The Actual Cmd Line Calls
#

createGroup() {
	shout "Create Azure group"
	date

	az group create \
	  --name "${RS_GROUP}" \
	  --location "${REGION}"

	# Wait until resource group will be visible in azure.
	counter=0
	until [[ $(az group exists --name "${RS_GROUP}" -o json) == true ]]; do
		sleep 15
		counter=$(( counter + 1 ))
		if (( counter == 5 )); then
			echo -e "---\nAzure resource group ${RS_GROUP} still not present after one minute wait.\n---"
			exit 1
		fi
	done
}

# Create The Azure EventHubs Namespace Based On Global Configuration
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

# Get The Azure EventHub Namespace Authorization Keys' PrimaryConnectionString
cmdNamespacePrimaryConnectionString() {
  az eventhubs namespace authorization-rule keys list \
    -o json \
    --resource-group "${RS_GROUP}" \
    --namespace-name "${EVENTHUB_NAMESPACE_NAME}" \
    --name "${EVENTHUB_NAMESPACE_SHARED_ACCESS_KEY}" | \
    jq -r ".primaryConnectionString"
}


# Create The EventHub Namespace Secret And Return It
cmdCreateventHubNamespaceSecret() {

  IFS="" read -r -d '' kafkaSecret << EOF
apiVersion: v1
stringData:
  kafka.brokers: ${K8S_SECRET_BROKER}
  kafka.namespace: ${EVENTHUB_NAMESPACE_NAME}
  kafka.password: ${K8S_SECRET_PASSWORD}
  kafka.username: ${K8S_SECRET_USERNAME}
  kafka.secretName: knative-kafka
  environment.kafkaProvider: azure
kind: Secret
metadata:
  name: ${K8S_SECRET_NAME}
  namespace: ${K8S_SECRET_NAMESPACE}
  labels:
    knativekafka.kyma-project.io/kafka-secret: "true"
    installer: overrides
    component: knative-eventing-channel-kafka
    kyma-project.io/installation: ""
type: Opaque
EOF
  echo "$kafkaSecret"
}

# Verify The Expected Dependencies Are Present On $PATH
verifyPathDependencies() {

  if ! [[ -x "$(command -v az)" ]]; then
    shout "Executable 'az' Not Found On \$PATH - Exiting"
    exit 1
  fi

  if ! [[ -x "$(command -v jq)" ]]; then
    shout "Executable 'jq' Not Found On \$PATH - Exiting"
    exit 1
  fi

}


function azureAuthenticating() {
	shout "Authenticating to azure"
	date

	az login \
	  --service-principal \
	  -u "${AZURE_SUBSCRIPTION_APP_ID}" \
	  -p "${AZURE_SUBSCRIPTION_SECRET}" \
	  --tenant "${AZURE_SUBSCRIPTION_TENANT}"
	az account set \
	  --subscription "${AZURE_SUBSCRIPTION_ID}"
}

# Enable this while debugging to confirm The User's Desire To Provision
# A New EventHub Namespace For Their Current Azure/K8S Context
confirmConfiguration() {

  # Log The Configuration Summary
  shout "The following configuration will be used to provision the new EventHub Namespace - review for correctness before continuing!"
  shout "Azure Subscription: ${AZURE_SUBSCRIPTION}"
  shout "Azure Resource Group: ${RS_GROUP}"
  shout "New EventHub Namespace name: ${EVENTHUB_NAMESPACE_NAME}"
  shout "New EventHub Namespace location: ${EVENTHUB_NAMESPACE_LOCATION}"
  shout "New EventHub Namespace throughput min: ${EVENTHUB_NAMESPACE_MIN_THROUGHPUT_UNITS}"
  shout "New EventHub Namespace throughput max: ${EVENTHUB_NAMESPACE_MAX_THROUGHPUT_UNITS}"
  shout "Kubernetes Secret name: ${EVENTHUB_NAMESPACE_NAME}"
  shout "Kubernetes Secret Namespace: ${K8S_SECRET_NAMESPACE}"
}


# Create The EventHub Namespace Based On User's Current Azure / K8S Context
createEventHubNamespace() {

  # Execute The Azure EventHubs Namespace Creation Command & Handle The Results
  shout "Creating New EventHubs Namespace... (takes several minutes - be patient :)"
  if [[ $(cmdCreateEventHubNamespace) -eq 0 ]]; then
    shout "Successfully Created New EventHub Namespace!"
  else
    shout "Failed To Create New EventHub Namespace - Exiting!"
    exit 1
  fi
}


# Load The EventHub Namespace's Authorization Key Information Into Global Variables
loadAuthorizationKey() {

  # Get The New EventHub Namespace's PrimaryConnectionString
  shout "Loading The New EventHub Namespace's Authorization Key..."
  local primaryConnectionString=""
  primaryConnectionString=$(cmdNamespacePrimaryConnectionString)

  # Populate The Kubernetes Secret Broker / Password Values
  K8S_SECRET_BROKER=$(echo "${primaryConnectionString}" | sed -e "s/^Endpoint=.*sb:\/\/\(.*\)\/;.*$/\1:${KAFKA_BROKER_PORT}/")
  K8S_SECRET_PASSWORD=${primaryConnectionString}
}


# Print The EventHub Namespace Secret (Contains The EventHub Namespace Auth Keys)
createK8SSecretFile() {

  shout "Creating a Kubernetes Secret override For The New EventHub Namespace..."
  eventHubSecret=$(cmdCreateventHubNamespaceSecret)
  echo "$eventHubSecret" >> installer-config-azure-eventhubs.yaml.tpl
}

#
# Main Script Execution
#

# Verify The Environment Contains The Expected Dependencies (az, kubectl, jq, etc.)
verifyPathDependencies

# Confirm the configuration
confirmConfiguration

# Authenticating in Azure
azureAuthenticating

# Create The New Azure Resource Group
createGroup

# Create The New Azure EventHubs Namespace
createEventHubNamespace

# Lookup The New Azure EventHub Namespace's Authorization Key
loadAuthorizationKey

# Create K8S Secret For EventHub Namespace
createK8SSecretFile
