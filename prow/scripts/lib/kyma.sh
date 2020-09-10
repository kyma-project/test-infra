#!/usr/bin/env bash

# kyma::install starts Kyma installation on the cluster
#
# Arguments:
#   $1 - Path to the Kyma sources
#   $2 - Name of the installer Docker Image
#   $3 - Path to the installer resource
#   $4 - Path to the installer custom resource
#   $5 - Installation timeout
function kyma::install {
    "${1}/installation/scripts/concat-yamls.sh" "${3}" "${4}" \
        | sed -e 's;image: eu.gcr.io/kyma-project/.*installer:.*$;'"image: ${2};" \
        | sed -e "s/__VERSION__/0.0.1/g" \
        | sed -e "s/__.*__//g" \
        | kubectl apply -f-
    
    kyma::is_installed "${1}" "${5}"
}

# kyma::is_installed waits for Kyma installation finish
#
# Arguments:
#   $1 - Path to the Kyma sources
#   $2 - Installation timeout
function kyma::is_installed {
    "${1}/installation/scripts/is-installed.sh" --timeout "${2}"
}

# kyma::load_config loads Kyma overrides to the cluster. Also sets domain and cluster IP
#
# Arguments:
#   $1 - IP of the cluster
#   $2 - Domain name
#   $3 - Path to the overrides file
function kyma::load_config {
    kubectl create namespace "kyma-installer" || echo "Ignore namespace creation"

    < "${3}" sed 's/\.minikubeIP: .*/\.minikubeIP: '"${1}"'/g' \
        | sed 's/\.domainName: .*/\.domainName: '"${2}"'/g' \
        | kubectl apply -f-
}

# kyma::test starts the Kyma integration tests
#
# Arguments:
#   $1 - Path to the scripts (../) directory
function kyma::test {
    "${1}/kyma-testing.sh"
}

# kyma::build_installer builds Kyma Installer Docker image
#
# Arguments:
#   $1 - Path to the Kyma sources
#   $2 - Installer Docker image name
function kyma::build_installer {
    docker build -t "${2}" -f "${1}/tools/kyma-installer/kyma.Dockerfile" "${1}"
}

# kyma::build_and_push_installer builds Kyma Installer Docker image
#
# Arguments:
#   $1 - Path to the Kyma sources
#   $2 - Installer Docker image name
function kyma::build_and_push_installer {
    docker build -t "${2}" -f "${1}/tools/kyma-installer/kyma.Dockerfile" "${1}"
    docker push "${2}"
}

# kyma::update_hosts appends hosts file with Kyma DNS records
function kyma::update_hosts {
    # TODO(michal-hudy):  Switch to local DNS server if possible
    local -r hosts="$(kubectl get virtualservices --all-namespaces -o jsonpath='{.items[*].spec.hosts[*]}')"
    echo "127.0.0.1 ${hosts}" | tee -a /etc/hosts > /dev/null
}

# kyma::get_last_release_version returns latest Kyma release version
#
# Arguments:
#   $1 - GitHub token
# Returns:
#   Last Kyma release version
function kyma::get_last_release_version {
    version=$(curl --silent --fail --show-error "https://api.github.com/repos/kyma-project/kyma/releases?access_token=${1}" \
        | jq -r 'del( .[] | select( (.prerelease == true) or (.draft == true) )) | sort_by(.tag_name | split(".") | map(tonumber)) | .[-1].tag_name')

    echo "${version}"
}
