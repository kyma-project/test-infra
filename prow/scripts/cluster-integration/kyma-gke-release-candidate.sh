#!/usr/bin/env bash

#Description: Kyma with central connector-service plan on GKE. This scripts implements a pipeline that consists of many steps. The purpose is to install and test Kyma on real GKE cluster with central connector-service.
#
#
#Expected vars:
#
# - REPO_OWNER - Set up by prow, repository owner/organization
# - REPO_NAME - Set up by prow, repository name
# - KYMA_PROJECT_DIR - directory path with Kyma sources to use for installation
# - CLOUDSDK_CORE_PROJECT - GCP project for all GCP resources used during execution (Service Account, IP Address, DNS Zone, image registry etc.)
# - CLOUDSDK_COMPUTE_ZONE - GCP compute zone
# - CLOUDSDK_COMPUTE_REGION - GCP compute region
# - CLOUDSDK_DNS_ZONE_NAME - GCP zone name (not its DNS name!)
# - GOOGLE_APPLICATION_CREDENTIALS - GCP Service Account key file path
# - GKE_CLUSTER_VERSION - GKE cluster version
# - MACHINE_TYPE - (optional) GKE machine type
#
#Permissions: In order to run this script you need to use a service account with permissions equivalent to the following GCP roles:
# - Compute Admin
# - Kubernetes Engine Admin
# - Kubernetes Engine Cluster Admin
# - DNS Administrator
# - Service Account User
# - Storage Admin
# - Compute Network Admin

set -o errexit

#Exported variables
export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"
export TTL_HOURS=168 #7 days

# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/docker.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/docker.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "$TEST_INFRA_SOURCES_DIR/prow/scripts/lib/gcp.sh"
# shellcheck source=prow/scripts/lib/kyma.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"

requiredVars=(
    REPO_OWNER
    REPO_NAME KYMA_PROJECT_DIR
    CLOUDSDK_CORE_PROJECT
    CLOUDSDK_COMPUTE_REGION
    CLOUDSDK_DNS_ZONE_NAME
    GOOGLE_APPLICATION_CREDENTIALS
    GKE_CLUSTER_VERSION
    CERTIFICATES_BUCKET
)

utils::check_required_vars "${requiredVars[@]}"

trap cleanupOnError EXIT INT

#!Put cleanup code in this function! Function is executed at !!!exit from the script!!! and on interuption.
cleanupOnError() {
    #!!! Must be at the beginning of this function !!!
    EXIT_STATUS=$?

    # Do not cleanup cluster if job finished successfully
    if [ "$EXIT_STATUS" == "0" ] ; then
        log::info "Job finished successfully, cleanup will not be performed"
        exit
    fi

    if [ "${ERROR_LOGGING_GUARD}" = "true" ]; then
        log::error "AN ERROR OCCURED! Take a look at preceding log entries."
        echo
    fi

    #Turn off exit-on-error so that next step is executed even if previous one fails.
    set +e

    if [[ "${CLEANUP_CLUSTER}" == "true" ]] ; then
        log::info "Deprovision cluster: \"${COMMON_NAME}\""

        #save disk names while the cluster still exists to remove them later
        DISKS=$(kubectl get pvc --all-namespaces -o jsonpath="{.items[*].spec.volumeName}" | xargs -n1 echo)
        export DISKS

        #Delete cluster
        gcp::deprovision_k8s_cluster \
            -n "$COMMON_NAME" \
            -p "$CLOUDSDK_CORE_PROJECT" \
            -z "$CLOUDSDK_COMPUTE_ZONE" \
        #Delete orphaned disks
        log::info "Delete orphaned PVC disks..."
        for NAMEPATTERN in ${DISKS}
        do
            DISK_NAME=$(gcloud compute disks list --filter="name~${NAMEPATTERN}" --format="value(name)")
            echo "Removing disk: ${DISK_NAME}"
            gcloud compute disks delete "${DISK_NAME}" --quiet
        done
    fi

    if [ -n "${CLEANUP_GATEWAY_DNS_RECORD}" ]; then
        log::info "Delete Gateway DNS Record"
        gcp::delete_dns_record \
            -a "$GATEWAY_IP_ADDRESS" \
            -h "*" \
            -s "$COMMON_NAME" \
            -p "$CLOUDSDK_CORE_PROJECT" \
            -z "$CLOUDSDK_DNS_ZONE_NAME"
    fi

    if [ -n "${CLEANUP_GATEWAY_IP_ADDRESS}" ]; then
        gcp::delete_ip_address \
            -n "${GATEWAY_IP_ADDRESS_NAME}" \
            -p "$CLOUDSDK_CORE_PROJECT" \
            -R "$CLOUDSDK_COMPUTE_REGION"
    fi

    MSG=""
    if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
    log::info "Job is finished ${MSG}"
    date
    set -e

    exit "${EXIT_STATUS}"
}

installKyma() {

	kymaUnsetVar=false

  # shellcheck disable=SC2043
	for var in GATEWAY_IP_ADDRESS ; do
    	if [ -z "${!var}" ] ; then
        	echo "ERROR: $var is not set"
        	kymaUnsetVar=true
    	fi
	done
	if [ "${kymaUnsetVar}" = true ] ; then
    	exit 1
	fi

	CERTIFICATES_BUCKET="${CERTIFICATES_BUCKET}" "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/get-letsencrypt-cert.sh"
	TLS_CERT=$(base64 -i ./letsencrypt/live/"${DOMAIN}"/fullchain.pem | tr -d '\n')
	export TLS_CERT
	TLS_KEY=$(base64 -i ./letsencrypt/live/"${DOMAIN}"/privkey.pem   | tr -d '\n')
	export TLS_KEY

	log::info "Trigger installation"
	set -x

	kyma deploy \
			--ci \
			--source local \
			--workspace "${KYMA_SOURCES_DIR}" \
			--domain "${DOMAIN}" \
			--profile production \
			--tls-crt "./letsencrypt/live/${DOMAIN}/fullchain.pem" \
			--tls-key "./letsencrypt/live/${DOMAIN}/privkey.pem" \
			--value "istio.components.ingressGateways.config.service.loadBalancerIP=${GATEWAY_IP_ADDRESS}" \
			--value "global.domainName=${DOMAIN}" \
			--timeout 60m

	set +x

}

# Enforce lowercase
readonly REPO_OWNER=$(echo "${REPO_OWNER}" | tr '[:upper:]' '[:lower:]')
export REPO_OWNER
readonly REPO_NAME=$(echo "${REPO_NAME}" | tr '[:upper:]' '[:lower:]')
export REPO_NAME

### Cluster name must be less than 40 characters!
readonly COMMON_NAME_PREFIX="gke-release"
readonly RELEASE_VERSION=$(cat "VERSION")
log::info "Reading release version from RELEASE_VERSION file, got: ${RELEASE_VERSION}"
TRIMMED_RELEASE_VERSION=${RELEASE_VERSION//./-}
COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${TRIMMED_RELEASE_VERSION}" | tr "[:upper:]" "[:lower:]")

gcp::set_vars_for_network \
  -n "$JOB_NAME"
export GCLOUD_NETWORK_NAME="${gcp_set_vars_for_network_return_net_name:?}"
export GCLOUD_SUBNET_NAME="${gcp_set_vars_for_network_return_subnet_name:?}"

#Local variables
DNS_SUBDOMAIN="${COMMON_NAME}"

#Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"

log::info "Authenticate"
gcp::authenticate \
    -c "${GOOGLE_APPLICATION_CREDENTIALS}"
docker::start
DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --format="value(dnsName)")"
export DNS_DOMAIN
DOMAIN="${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"
export DOMAIN

log::info "Reserve IP Address for Ingressgateway"
GATEWAY_IP_ADDRESS_NAME="${COMMON_NAME}"
gcp::reserve_ip_address \
    -n "${GATEWAY_IP_ADDRESS_NAME}" \
    -p "$CLOUDSDK_CORE_PROJECT" \
    -r "$CLOUDSDK_COMPUTE_REGION"
GATEWAY_IP_ADDRESS="${gcp_reserve_ip_address_return_ip_address:?}"
CLEANUP_GATEWAY_IP_ADDRESS="true"
echo "Created IP Address for Ingressgateway: ${GATEWAY_IP_ADDRESS}"


log::info "Create DNS Record for Ingressgateway IP"
gcp::create_dns_record \
    -a "$GATEWAY_IP_ADDRESS" \
    -h "*" \
    -s "$COMMON_NAME" \
    -p "$CLOUDSDK_CORE_PROJECT" \
    -z "$CLOUDSDK_DNS_ZONE_NAME"
CLEANUP_GATEWAY_DNS_RECORD="true"

log::info "Create ${GCLOUD_NETWORK_NAME} network with ${GCLOUD_SUBNET_NAME} subnet"
gcp::create_network \
    -n "${GCLOUD_NETWORK_NAME}" \
    -s "${GCLOUD_SUBNET_NAME}" \
    -p "$CLOUDSDK_CORE_PROJECT"

log::info "Provision cluster: \"${COMMON_NAME}\""
export GCLOUD_SERVICE_KEY_PATH="${GOOGLE_APPLICATION_CREDENTIALS}"

gcp::provision_k8s_cluster \
        -c "$COMMON_NAME" \
        -p "$CLOUDSDK_CORE_PROJECT" \
        -v "$GKE_CLUSTER_VERSION" \
        -j "$JOB_NAME" \
        -J "$PROW_JOB_ID" \
        -t "$TTL_HOURS" \
        -z "$CLOUDSDK_COMPUTE_ZONE" \
        -R "$CLOUDSDK_COMPUTE_REGION" \
        -N "$GCLOUD_NETWORK_NAME" \
        -S "$GCLOUD_SUBNET_NAME" \
        -g "$GCLOUD_SECURITY_GROUP_DOMAIN" \
        -P "$TEST_INFRA_SOURCES_DIR" \
        -r "$PROVISION_REGIONAL_CLUSTER" \
        -m "$MACHINE_TYPE" \
        -D "$CLUSTER_USE_SSD"
CLEANUP_CLUSTER="true"

kyma::install_cli

log::info "Install kyma"
installKyma

#---
log::info "create cluster-admin binding for kyma_developers@sap.com"
kubectl create clusterrolebinding kyma_developers --clusterrole cluster-admin --group kyma_developers@sap.com
# TODO (@Ressetkk): Move this part as re-usable function if needed
log::info "generate service account and kubeconfig with cluster-admin rights"
namespace="default"
serviceAccount="admin-user"
secretName="$serviceAccount-secret"

server="https://$(gcloud container clusters describe "$COMMON_NAME" --region "$CLOUDSDK_COMPUTE_REGION" | awk '/endpoint:/ {print $2}')"
kubectl create serviceaccount -n "$namespace" "$serviceAccount"
kubectl apply -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: "$secretName"
  namespace: "$namespace"
  annotations:
    kubernetes.io/service-account.name: "$serviceAccount"
type: kubernetes.io/service-account-token
EOF
kubectl create clusterrolebinding $serviceAccount --clusterrole cluster-admin --serviceaccount="$namespace:$serviceAccount"

ca="$(kubectl -n "$namespace" get "secret/$secretName" -o jsonpath='{.data.ca\.crt}')"
token="$(kubectl -n "$namespace" get "secret/$secretName" -o jsonpath='{.data.token}' | base64 --decode)"

echo "---
apiVersion: v1
kind: Config
clusters:
  - name: default
    cluster:
      certificate-authority-data: ${ca}
      server: ${server}
contexts:
  - name: default
    context:
      cluster: default
      namespace: default
      user: ${serviceAccount}
users:
  - name: ${serviceAccount}
    user:
      token: ${token}
current-context: default
" > kubeconfig

# escape kubeconfig properly
pubsub_message=$(jq -c --null-input "{\"cluster_name\": \"${COMMON_NAME}\", \"kyma_version\": \"${RELEASE_VERSION}\", \"kubeconfig\": \"$(cat kubeconfig)\"}")
gcloud pubsub topics publish --project="${PUBSUB_PROJECT}" "${PUBSUB_TOPIC}" --message="${pubsub_message}"
#---

log::info "Collect list of images"
if [ -z "$ARTIFACTS" ] ; then
    ARTIFACTS=/tmp/artifacts
fi

IMAGES_LIST=$(kubectl get pods --all-namespaces -o go-template --template='{{range .items}}{{range .status.containerStatuses}}{{.name}},{{.image}},{{.imageID}}{{printf "\n"}}{{end}}{{range .status.initContainerStatuses}}{{.name}},{{.image}},{{.imageID}}{{printf "\n"}}{{end}}{{end}}' | uniq | sort)
echo "${IMAGES_LIST}" > "${ARTIFACTS}/kyma-images-release-${RELEASE_VERSION}.csv"

# also generate image list in json
## this is false-positive as we need to use single-quotes for jq
# shellcheck disable=SC2016
IMAGES_LIST=$(kubectl get pods --all-namespaces -o json | jq '{ images: [.items[] | .metadata.ownerReferences[0].name as $owner | (.status.containerStatuses + .status.initContainerStatuses)[] | { name: .imageID, custom_fields: {owner: $owner, image: .image, name: .name }}] | unique | group_by(.name) | map({name: .[0].name, custom_fields: {owner: map(.custom_fields.owner) | unique | join(","), container_name: map(.custom_fields.name) | unique | join(","), image: .[0].custom_fields.image}}) | map(select (.name | startswith("sha256") | not))}' )
echo "${IMAGES_LIST}" > "${ARTIFACTS}/kyma-images-release-${RELEASE_VERSION}.json"

log::success "Success"

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
