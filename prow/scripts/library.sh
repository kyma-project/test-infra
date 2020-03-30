#!/usr/bin/env bash

# DEPRECATED - use scripts from `lib` directory

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

LICENSE_PULLER_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/license-puller.sh"
export LICENSE_PULLER_PATH

function start_docker() {
    echo "Docker in Docker enabled, initializing..."
    printf '=%.0s' {1..80}; echo
    # If we have opted in to docker in docker, start the docker daemon,
    service docker start
    # the service can be started but the docker socket not ready, wait for ready
    local WAIT_N=0
    local MAX_WAIT=20
    while true; do
        # docker ps -q should only work if the daemon is ready
        docker ps -q > /dev/null 2>&1 && break
        if [[ ${WAIT_N} -lt ${MAX_WAIT} ]]; then
            WAIT_N=$((WAIT_N+1))
            echo "Waiting for docker to be ready, sleeping for ${WAIT_N} seconds."
            sleep ${WAIT_N}
        else
            echo "Reached maximum attempts, not waiting any longer..."
            exit 1
        fi
    done
    printf '=%.0s' {1..80}; echo

    docker-credential-gcr configure-docker
    echo "Done setting up docker in docker."
}

function authenticate() {
    echo "Authenticating"
    gcloud auth activate-service-account --key-file "${GOOGLE_APPLICATION_CREDENTIALS}" || exit 1

}

function authenticateDocker() {
    shout "Authenticating on docker registry ${DOCKER_REGISTRY}"

    gcloud auth print-access-token | docker login -u oauth2accesstoken --password-stdin https://"${DOCKER_REGISTRY}"
}

function configure_git() {
    echo "Configuring git"
    # configure ssh
    if [[ ! -z "${BOT_GITHUB_SSH_PATH}" ]]; then
        mkdir "${HOME}/.ssh/"
        cp "${BOT_GITHUB_SSH_PATH}" "${HOME}/.ssh/ssh_key.pem"
        local SSH_FILE="${HOME}/.ssh/ssh_key.pem"
        touch "${HOME}/.ssh/known_hosts"
        ssh-keyscan -H github.com >> "${HOME}/.ssh/known_hosts"
        chmod 400 "${SSH_FILE}"
        eval "$(ssh-agent -s)"
        ssh-add "${SSH_FILE}"
        ssh-add -l
        git config --global core.sshCommand "ssh -i ${SSH_FILE}"
    fi

    # configure email
    if [[ ! -z "${BOT_GITHUB_EMAIL}" ]]; then
        git config --global user.email "${BOT_GITHUB_EMAIL}"
    fi

    # configure name
    if [[ ! -z "${BOT_GITHUB_NAME}" ]]; then
        git config --global user.name "${BOT_GITHUB_NAME}"
    fi
}

function init() {
    echo "Initializing"

    if [[ ! -z "${GOOGLE_APPLICATION_CREDENTIALS}" ]]; then
        authenticate
    fi

    if [[ "${DOCKER_IN_DOCKER_ENABLED}" == true ]]; then
        start_docker
    fi

	if [[ "${DOCKER_IN_DOCKER_ENABLED}" == true ]] && [[ "${PERFORMACE_CLUSTER_SETUP}" == "true" ]]; then
	    authenticateDocker
	fi

    if [[ ! -z "${BOT_GITHUB_SSH_PATH}" ]] || [[ ! -z "${BOT_GITHUB_EMAIL}" ]] || [[ ! -z "${BOT_GITHUB_NAME}" ]]; then
        configure_git
    fi
}

function shout() {
    echo -e "${GREEN}
#################################################################################################
# $1
#################################################################################################
    ${NC}"
}

function shoutFail() {
    echo -e "${RED}
#################################################################################################
# $1
#################################################################################################
    ${NC}"
}

function checkInputParameterValue() {
    if [ -z "${1}" ] || [ "${1:0:2}" == "--" ]; then
        echo -e "${RED}Wrong parameter value"
        echo -e "${RED}Make sure parameter value is neither empty nor start with two hyphens"
        exit 1
    fi
}

function checkClusterGradeInputParameterValue() {
    if [[  "${CLUSTER_GRADE}" != "production" ]] && [[ "${CLUSTER_GRADE}" != "development" ]]; then
        shoutFail "--cluster-grade  possible values are 'production' or 'development'"
        exit 1
    fi
}

function checkActionInputParameterValue() {
    if [[ "${ACTION}" != "create" ]] && [[ "${ACTION}" != "delete" ]]; then
        shoutFail "--action  possible values are 'create' or 'delete'"
        exit 1
    fi
}

function checkInfraInputParameterValue() {
    if [[ "${INFRA}" != "aks" ]] && [[ "${ACTION}" != "gke" ]]; then
        shoutFail "--infra  possible values are 'aks' or 'gke'"
        exit 1
    fi
}

function applyDexGithubConnectorOverride() {
	shout "Apply Dex Githubauth connector overrides"
	export DEX_CALLBACK_URL="https://dex.${DOMAIN}/callback"
	
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: dex-config-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: dex
    kyma-project.io/installation: ""
data:
 connectors: |
  - type: github
    id: github
    name: GitHub
    config:
      clientID: ${GITHUB_INTEGRATION_APP_CLIENT_ID}
      clientSecret: ${GITHUB_INTEGRATION_APP_CLIENT_SECRET}
      redirectURI: ${DEX_CALLBACK_URL}
      orgs:
      - name: kyma-project
EOF
}

function applyDexGithibKymaAdminGroup() {
    kubectl get ClusterRoleBinding kyma-admin-binding -oyaml > kyma-admin-binding.yaml && cat >> kyma-admin-binding.yaml <<EOF 
- apiGroup: rbac.authorization.k8s.io
  kind: Group
  name: kyma-project:cluster-access
EOF

    kubectl replace -f kyma-admin-binding.yaml
}

#Update stackdriver-metadata-agent memory settings
function updatememorysettings() {

cat <<EOF | kubectl replace -f -
apiVersion: v1
data:
  NannyConfiguration: |-
    apiVersion: nannyconfig/v1alpha1
    kind: NannyConfiguration
    baseMemory: 100Mi
kind: ConfigMap
metadata:
  labels:
    addonmanager.kubernetes.io/mode: EnsureExists
    kubernetes.io/cluster-service: "true"
  name: metadata-agent-config
  namespace: kube-system
EOF

	kubectl delete deployment -n kube-system stackdriver-metadata-agent-cluster-level

}
