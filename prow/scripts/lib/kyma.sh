#!/usr/bin/env bash

function kyma::install {
    # TODO: (@michal-hudy) It probably works only with Kyma Lite, but I'm not sure, will be improved during full migration
    "${1}/installation/scripts/concat-yamls.sh" "${3}" "${4}" \
        | sed -e 's;image: eu.gcr.io/kyma-project/.*installer:.*$;'"image: ${2};" \
        | sed -e "s/__VERSION__/0.0.1/g" \
        | sed -e "s/__.*__//g" \
        | kubectl apply -f-
    
    # TODO: Szostok wbijaj tutaj :)
    "${1}/installation/scripts/is-installed.sh" --timeout "${5}"
}

function kyma::load_config {
    kubectl create namespace "kyma-installer"

    < "${3}" sed 's/\.minikubeIP: .*/\.minikubeIP: '"${1}"'/g' \
        | sed 's/\.domainName: .*/\.domainName: '"${2}"'/g' \
        | kubectl apply -f-
}

function kyma::test {
    # TODO: Szostok wbijaj tutaj :)
    "${1}/installation/scripts/testing.sh" --cleanup "false" --concurrency 5
}

function kyma::build_installer {
    docker build -t "${2}" -f "${1}/tools/kyma-installer/kyma.Dockerfile" "${1}"
}

function kyma::update_hosts {
    # TODO: (@michal-hudy) Switch to local DNS server if possible
    local -r hosts="$(kubectl get virtualservices --all-namespaces -o jsonpath='{.items[*].spec.hosts[*]}')"
    echo "127.0.0.1 ${hosts}" | tee -a /etc/hosts > /dev/null
}

function kyma::install_tiller {
    "${1}/installation/scripts/install-tiller.sh"
}