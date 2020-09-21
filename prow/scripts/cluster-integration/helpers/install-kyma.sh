#!/usr/bin/env bash

#Description: Installs Kyma in a given GKE cluster
#
#Expected vars:
# - GATEWAY_IP_ADDRESS: static IP for gateway
# - DOCKER_PUSH_REPOSITORY: name of the docker registry where images are pushed
# - KYMA_SOURCES_DIR: absolute path for kyma sources directory
# - DOCKER_PUSH_DIRECTORY: directory for docker images where it should be pushed
# - GOOGLE_APPLICATION_CREDENTIALS - GCP Service Account key file path
# - STANDARIZED_NAME: a variation of cluster name
# - REPO_OWNER: Kyma repository owner
# - REPO_NAME: name of the Kyma repository
# - CURRENT_TIMESTAMP: Current timestamp which is computed as $(date +%Y%m%d)
# - DOMAIN: Combination of gcloud managed-zones and cluster name "${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"

# shellcheck disable=SC1090
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/library.sh"

function installKyma() {

    kymaUnsetVar=false

    for var in DOCKER_PUSH_REPOSITORY KYMA_SOURCES_DIR DOCKER_PUSH_DIRECTORY GOOGLE_APPLICATION_CREDENTIALS STANDARIZED_NAME REPO_OWNER REPO_NAME CURRENT_TIMESTAMP GCR_PUSH_GOOGLE_APPLICATION_CREDENTIALS; do
        if [ -z "${!var}" ] ; then
            echo "ERROR: $var is not set"
            kymaUnsetVar=true
        fi
    done

    if [[ "${PERFORMACE_CLUSTER_SETUP}" == "" ]]; then
        for var in GATEWAY_IP_ADDRESS DOMAIN; do
            if [ -z "${!var}" ] ; then
                echo "ERROR: $var is not set"
                kymaUnsetVar=true
            fi
        done
    fi

    if [ "${kymaUnsetVar}" = true ] ; then
        exit 1
    fi

    # shellcheck disable=SC2153
    KYMA_RESOURCES_DIR="${KYMA_SOURCES_DIR}/installation/resources"
    INSTALLER_YAML="${KYMA_RESOURCES_DIR}/installer.yaml"
    INSTALLER_CR="${KYMA_RESOURCES_DIR}/installer-cr-cluster.yaml.tpl"

    if [[ "${PERFORMACE_CLUSTER_SETUP}" == "" ]]; then
        export KYMA_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}/${STANDARIZED_NAME}/${REPO_OWNER}/${REPO_NAME}:${CURRENT_TIMESTAMP}"
    else
        export KYMA_INSTALLER_IMAGE="${DOCKER_PUSH_REPOSITORY}${DOCKER_PUSH_DIRECTORY}:${CURRENT_TIMESTAMP}"
    fi

    shout "Build Kyma-Installer Docker image"
    date

    createImage

    shout "Apply Kyma config"
    date
    sed -e 's;image: eu.gcr.io/kyma-project/.*/installer:.*$;'"image: ${KYMA_INSTALLER_IMAGE};" "${INSTALLER_YAML}" \
    | kubectl apply -f-

    if [[ "${PERFORMACE_CLUSTER_SETUP}" == "" ]]; then
        # shellcheck disable=SC1090
        source "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}"/generate-and-export-letsencrypt-TLS-cert.sh

        cat << EOF > $PWD/istio-overrides
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
spec:
  components:
    ingressGateways:
      - name: istio-ingressgateway
        k8s:
          service:
            loadBalancerIP: ${GATEWAY_IP_ADDRESS}
            type: LoadBalancer
EOF

        "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "istio-overrides" \
            --label "component=istio" \
            --file "$PWD/istio-overrides"

        "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "installation-config-overrides" \
            --data "global.domainName=${DOMAIN}" \
            --data "global.loadBalancerIP=${GATEWAY_IP_ADDRESS}"

        "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "cluster-certificate-overrides" \
            --data "global.tlsCrt=${TLS_CERT}" \
            --data "global.tlsKey=${TLS_KEY}"

    fi

    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "core-test-ui-acceptance-overrides" \
        --data "test.acceptance.ui.logging.enabled=true" \
        --label "component=core"

    "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-config-map.sh" --name "application-registry-overrides" \
        --data "application-registry.deployment.args.detailedErrorResponse=true" \
        --label "component=application-connector"

    waitUntilInstallerApiAvailable

    if [[ "${PERFORMACE_CLUSTER_SETUP}" != "" ]]; then
        kubectl config set-context "gke_${CLOUDSDK_CORE_PROJECT}_${CLOUDSDK_COMPUTE_ZONE}_${INPUT_CLUSTER_NAME}" --namespace=default
    fi

    shout "Trigger installation"
    date

    sed -e "s/__VERSION__/0.0.1/g" "${INSTALLER_CR}"  | sed -e "s/__.*__//g" | kubectl apply -f-
    "${KYMA_SCRIPTS_DIR}"/is-installed.sh --timeout 30m
}

function createImage() {
    shout "Kyma Installer Image: ${KYMA_INSTALLER_IMAGE}"
    # shellcheck disable=SC1090
    source "${TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS}/create-image.sh"
}

function waitUntilInstallerApiAvailable() {
    shout "Waiting for Installer API"

    attempts=5
    for ((i=1; i<=attempts; i++)); do
        numberOfLines=$(kubectl api-versions | grep -c "installer.kyma-project.io")
        if [[ "$numberOfLines" == "1" ]]; then
            echo "API found"
            break
        elif [[ "${i}" == "${attempts}" ]]; then
            echo "ERROR: API not found, exit"
            exit 1
        fi

        echo "Sleep for 3 seconds"
        sleep 3
    done
}

installKyma
