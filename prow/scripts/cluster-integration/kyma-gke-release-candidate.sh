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
# - KYMA_ARTIFACTS_BUCKET: GCP bucket
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
# shellcheck source=prow/scripts/lib/gcloud.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gcloud.sh"
# shellcheck source=prow/scripts/lib/docker.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/docker.sh"

requiredVars=(
    REPO_OWNER
    REPO_NAME KYMA_PROJECT_DIR
    CLOUDSDK_CORE_PROJECT
    CLOUDSDK_COMPUTE_REGION
    CLOUDSDK_DNS_ZONE_NAME
    GOOGLE_APPLICATION_CREDENTIALS
    KYMA_ARTIFACTS_BUCKET
    GKE_CLUSTER_VERSION
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

    if [ -n "${CLEANUP_CLUSTER}" ]; then
        log::info "Deprovision cluster: \"${CLUSTER_NAME}\""

        #save disk names while the cluster still exists to remove them later
        DISKS=$(kubectl get pvc --all-namespaces -o jsonpath="{.items[*].spec.volumeName}" | xargs -n1 echo)
        export DISKS

        #Delete cluster
        gcloud::deprovision_gke_cluster "$CLUSTER_NAME"

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
        gcloud::delete_dns_record "${GATEWAY_IP_ADDRESS}" "${GATEWAY_DNS_FULL_NAME}"
    fi

    if [ -n "${CLEANUP_GATEWAY_IP_ADDRESS}" ]; then
        log::info "Release Gateway IP Address"
        gcloud::delete_ip_address "${GATEWAY_IP_ADDRESS_NAME}"
    fi

    if [ -n "${CLEANUP_APISERVER_DNS_RECORD}" ]; then
        log::info "Delete Apiserver proxy DNS Record"
        gcloud::delete_dns_record "${APISERVER_IP_ADDRESS}" "${APISERVER_DNS_FULL_NAME}"
    fi


    MSG=""
    if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
    log::info "Job is finished ${MSG}"
    date
    set -e

    exit "${EXIT_STATUS}"
}

# Enforce lowercase
readonly REPO_OWNER=$(echo "${REPO_OWNER}" | tr '[:upper:]' '[:lower:]')
export REPO_OWNER
readonly REPO_NAME=$(echo "${REPO_NAME}" | tr '[:upper:]' '[:lower:]')
export REPO_NAME

readonly COMMON_NAME_PREFIX="gke-release"
readonly RELEASE_VERSION=$(cat "VERSION")
log::info "Reading release version from RELEASE_VERSION file, got: ${RELEASE_VERSION}"
TRIMMED_RELEASE_VERSION=${RELEASE_VERSION//./-}
COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${TRIMMED_RELEASE_VERSION}" | tr "[:upper:]" "[:lower:]")


### Cluster name must be less than 40 characters!
export CLUSTER_NAME="${COMMON_NAME}"

export GCLOUD_NETWORK_NAME="${COMMON_NAME_PREFIX}-net"
export GCLOUD_SUBNET_NAME="${COMMON_NAME_PREFIX}-subnet"

### For gcloud::provision_gke_cluster
export GCLOUD_PROJECT_NAME="${CLOUDSDK_CORE_PROJECT}"
export GCLOUD_COMPUTE_ZONE="${CLOUDSDK_COMPUTE_ZONE}"

#Local variables
DNS_SUBDOMAIN="${COMMON_NAME}"
KYMA_SCRIPTS_DIR="${KYMA_SOURCES_DIR}/installation/scripts"


#Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"

log::info "Authenticate"
gcloud::authenticate "${GOOGLE_APPLICATION_CREDENTIALS}"
docker::start
DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --format="value(dnsName)")"

log::info "Reserve IP Address for Ingressgateway"
GATEWAY_IP_ADDRESS_NAME="${COMMON_NAME}"
GATEWAY_IP_ADDRESS=$(gcloud::reserve_ip_address "${GATEWAY_IP_ADDRESS_NAME}")
CLEANUP_GATEWAY_IP_ADDRESS="true"
echo "Created IP Address for Ingressgateway: ${GATEWAY_IP_ADDRESS}"


log::info "Create DNS Record for Ingressgateway IP"
GATEWAY_DNS_FULL_NAME="*.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
CLEANUP_GATEWAY_DNS_RECORD="true"
gcloud::create_dns_record "${GATEWAY_IP_ADDRESS}" "${GATEWAY_DNS_FULL_NAME}"


log::info "Create ${GCLOUD_NETWORK_NAME} network with ${GCLOUD_SUBNET_NAME} subnet"
gcloud::create_network "${GCLOUD_NETWORK_NAME}" "${GCLOUD_SUBNET_NAME}"

log::info "Provision cluster: \"${CLUSTER_NAME}\""
export GCLOUD_SERVICE_KEY_PATH="${GOOGLE_APPLICATION_CREDENTIALS}"
if [ -z "$MACHINE_TYPE" ]; then
      export MACHINE_TYPE="${DEFAULT_MACHINE_TYPE}"
fi
CLEANUP_CLUSTER="true"
gcloud::provision_gke_cluster "$CLUSTER_NAME"

log::info "Create CluserRoleBinding for ${GCLOUD_SECURITY_GROUP} group from ${GCLOUD_SECURITY_GROUP_DOMAIN} domain"
kubectl create clusterrolebinding kyma-developers-group-binding --clusterrole="cluster-admin" --group="${GCLOUD_SECURITY_GROUP}@${GCLOUD_SECURITY_GROUP_DOMAIN}"

log::info "Generate certificate"
DOMAIN="${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"

utils::generate_letsencrypt_cert "${DOMAIN}"

log::info "Apply Kyma config"

kubectl create namespace "kyma-installer"

"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "installation-config-overrides" \
    --data "global.domainName=${DOMAIN}" \
    --data "global.loadBalancerIP=${GATEWAY_IP_ADDRESS}"

"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "core-test-ui-acceptance-overrides" \
    --data "test.acceptance.ui.logging.enabled=true" \
    --label "component=core"

"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "cluster-certificate-overrides" \
    --data "global.tlsCrt=${TLS_CERT}" \
    --data "global.tlsKey=${TLS_KEY}"

cat << EOF > "$PWD/kyma_istio_operator"
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  namespace: istio-system
spec:
  components:
    ingressGateways:
      - name: istio-ingressgateway
        k8s:
          service:
            loadBalancerIP: ${GATEWAY_IP_ADDRESS}
            type: LoadBalancer
EOF

"${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map-file.sh" --name "istio-overrides" \
    --label "component=istio" \
    --file "$PWD/kyma_istio_operator"

echo "Use released artifacts"
    curl -L --silent --fail --show-error "https://github.com/kyma-project/kyma/releases/download/${RELEASE_VERSION}/kyma-installer.yaml" --output /tmp/kyma-installer.yaml
    curl -L --silent --fail --show-error "https://github.com/kyma-project/kyma/releases/download/${RELEASE_VERSION}/kyma-installer-cr-cluster.yaml" --output /tmp/kyma-installer-cr-cluster.yaml

\
# There is possibility of a race condition when applying kyma-installer-cluster.yaml
# Retry should prevent job from failing
n=0
until [ $n -ge 2 ]
do
    kubectl apply -f /tmp/kyma-installer.yaml && break
    echo "Failed to apply kyma-installer.yaml"
    n=$((n+1))
    if [ 2 -gt "$n" ]
    then
        echo "Retrying in 5 seconds"
        sleep 5
    else
        exit 1
    fi
done

n=0
until [ $n -ge 2 ]
do
    kubectl apply -f /tmp/kyma-installer-cr-cluster.yaml && break
    echo "Failed to apply kyma-installer-cr-cluster.yaml"
    n=$((n+1))
    if [ 2 -gt "$n" ]
    then
        echo "Retrying in 5 seconds"
        sleep 5
    else
        exit 1
    fi
done

log::info "Trigger installation"
"${KYMA_SCRIPTS_DIR}"/is-installed.sh --timeout 30m

if [ -n "$(kubectl get  service -n kyma-system apiserver-proxy-ssl --ignore-not-found)" ]; then
    log::info "Create DNS Record for Apiserver proxy IP"
    APISERVER_IP_ADDRESS=$(kubectl get  service -n kyma-system apiserver-proxy-ssl -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    APISERVER_DNS_FULL_NAME="apiserver.${DNS_SUBDOMAIN}.${DNS_DOMAIN}"
    CLEANUP_APISERVER_DNS_RECORD="true"
    gcloud::create_dns_record "${APISERVER_IP_ADDRESS}" "${APISERVER_DNS_FULL_NAME}"
fi

log::info "Collect list of images"
if [ -z "$ARTIFACTS" ] ; then
    ARTIFACTS:=/tmp/artifacts
fi

IMAGES_LIST=$(kubectl get pods --all-namespaces -o go-template --template='{{range .items}}{{range .status.containerStatuses}}{{.name}},{{.image}},{{.imageID}}{{printf "\n"}}{{end}}{{range .status.initContainerStatuses}}{{.name}},{{.image}},{{.imageID}}{{printf "\n"}}{{end}}{{end}}' | uniq | sort)
echo "${IMAGES_LIST}" > "${ARTIFACTS}/kyma-images-release-${RELEASE_VERSION}.csv"

# also generate image list in json
## this is false-positive as we need to use single-quotes for jq
# shellcheck disable=SC2016
IMAGES_LIST=$(kubectl get pods --all-namespaces -o json | jq '{ images: [.items[] | .metadata.ownerReferences[0].name as $owner | (.status.containerStatuses + .status.initContainerStatuses)[] | { name: .imageID, custom_fields: {owner: $owner, image: .image, name: .name }}] | unique | group_by(.name) | map({name: .[0].name, custom_fields: {owner: map(.custom_fields.owner) | unique | join(","), container_name: map(.custom_fields.name) | unique | join(","), image: .[0].custom_fields.image}})}' )
echo "${IMAGES_LIST}" > "${ARTIFACTS}/kyma-images-release-${RELEASE_VERSION}.json"

log::success "Success"

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
