#!/usr/bin/env bash

set -e

readonly driver="none"
readonly testsuiteName="testsuite-all"
KYMA_TEST_TIMEOUT=${KYMA_TEST_TIMEOUT:=1h}

date

host::create_registries_file(){
cat > registries.yaml <<EOL
mirrors:
  registry.localhost:5000:
    endpoint:
    - http://registry.localhost:5000
  configs: {}
  auths: {}

EOL
}

host::create_coredns_template(){

cat > coredns-patch.tpl <<EOL
data:
  Corefile: |
    registry.localhost:53 {
        hosts {
            REGISTRY_IP registry.localhost
        }
    }
    .:53 {
        errors
        health
        rewrite name regex (.*)\.local\.kyma\.dev istio-ingressgateway.istio-system.svc.cluster.local
        ready
        kubernetes cluster.local in-addr.arpa ip6.arpa {
          pods insecure
          upstream
          fallthrough in-addr.arpa ip6.arpa
        }
        hosts /etc/coredns/NodeHosts {
          reload 1s
          fallthrough
        }
        prometheus :9153
        forward . /etc/resolv.conf
        cache 30
        loop
        reload
        loadbalance
    }
EOL
}


host::create_docker_registry(){
# Create docker network
docker network create k3d-kyma

# Start docker Registry
docker run -d \
  -p 5000:5000 \
  --restart=always \
  --name registry.localhost \
  --network k3d-kyma \
  -v $PWD/registry:/var/lib/registry \
  registry:2
}


install::prereq(){
    curl https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash
    apt-get -y install jq
}

install::k3d() {
    echo "--> Installing k3d"
    curl -s https://raw.githubusercontent.com/rancher/k3d/main/install.sh | bash
    k3d version
    date
}

host::os() {
  local host_os
  case "$(uname -s)" in
    Darwin)
      host_os=darwin
      ;;
    Linux)
      host_os=linux
      ;;
    *)
      >&2 echo -e "Unsupported host OS. Must be Linux or Mac OS X."
      exit 1
      ;;
  esac
  echo "${host_os}"
}

install::kyma_cli() {
    local settings
    local kyma_version
    mkdir -p "/usr/local/bin"
    os=$(host::os)

    pushd "/usr/local/bin" || exit

    echo "Install kyma CLI ${os} locally to /usr/local/bin..."

    curl -sSLo kyma "https://storage.googleapis.com/kyma-cli-stable/kyma-${os}?alt=media"
    chmod +x kyma
    kyma_version=$(kyma version --client)
    echo "Kyma CLI version: ${kyma_version}"

    echo "OK"

    popd || exit

    eval "${settings}"
}

echo "--> Installing Kyma CLI"
if ! [[ -x "$(command -v kyma)" ]]; then
  echo "Kyma CLI not found"
  install::kyma_cli
else
  echo "Kyma CLI is already installed"
  kyma_version=$(kyma version --client)
  echo "Kyma CLI version: ${kyma_version}"
  minikube version
fi
echo "--> Done"

sed -i '1 s/localhost/localhost registry.localhost/' /etc/hosts

host::create_registries_file
host::create_coredns_template

host::create_docker_registry

install::prereq
install::k3d


date
echo "--> Creating k8s cluster"


#    --port 80:80@loadbalancer \
#    --port 443:443@loadbalancer \


k3d cluster create kyma \
    --network k3d-kyma \
    --volume $PWD/registries.yaml:/etc/rancher/k3s/registries.yaml \
    --k3s-server-arg --no-deploy \
    --k3s-server-arg traefik \
    --wait \
    --switch-context \
    --timeout 60s
date

while [[ $(kubectl get nodes -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do echo "Waiting for cluster nodes to be ready, elapsed time: $(( $SECONDS/60 )) min $(( $SECONDS % 60 )) sec"; sleep 2; done

export REGISTRY_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' /registry.localhost)
sed "s/REGISTRY_IP/$REGISTRY_IP/" coredns-patch.tpl >coredns-patch.yaml
kubectl -n kube-system patch cm coredns --patch "$(cat coredns-patch.yaml)"


#echo "--> Provision Kyma cluster on minikube using VM driver ${driver}"
#STARTTIME=$(date +%s)
#yes | kyma provision minikube \
#               --ci \
#               --vm-driver="${driver}"
#ENDTIME=$(date +%s)
#echo "  Execution time: $((ENDTIME - STARTTIME)) seconds."
#echo "--> Done"

echo "--> Installing Kyma on minikube cluster"

# This file will be created by cert-manager (not needed anymore):
#rm kyma/resources/core/charts/gateway/templates/kyma-gateway-certs.yaml
# apiserver-proxy dependencies are not required (cannot be disabled by values yet):
#rm kyma/resources/apiserver-proxy/requirements.yaml
#rm -R kyma/resources/apiserver-proxy/charts
#rm -R kyma/resources/iam-kubeconfig-service

STARTTIME=$(date +%s)
yes | kyma install \
     --ci \
     --source="local" \
     --src-path=./kyma \
     --custom-image="registry.localhost:5000/kyma-installer:master-f54ff69e"

#yes | kyma install \
#     --ci \
#     --source="eu.gcr.io/kyma-project/kyma-installer:master-f54ff69e" \
#     --src-path=./kyma


ENDTIME=$(date +%s)
echo "  Install time: $((ENDTIME - STARTTIME)) seconds."
echo "--> Done"

echo "--> Run kyma tests"
STARTTIME=$(date +%s)
echo "  List test definitions"
kyma test definitions --ci
echo "  Run tests"
kyma test run \
          --ci \
          --watch \
          --max-retries=1 \
          --name="${testsuiteName}" \
          --timeout="${KYMA_TEST_TIMEOUT}"

ENDTIME=$(date +%s)
echo "  Test time: $((ENDTIME - STARTTIME)) seconds."

echo "  Test summary"
kyma test status "${testsuiteName}" -owide
statusSucceeded=$(kubectl get cts "${testsuiteName}" -ojsonpath="{.status.conditions[?(@.type=='Succeeded')]}")
if [[ "${statusSucceeded}" != *"True"* ]]; then
  echo "- Fetching logs from testing pods in Failed status..."
  kyma test logs "${testsuiteName}" --test-status Failed

  echo "- Fetching logs from testing pods in Unknown status..."
  kyma test logs "${testsuiteName}" --test-status Unknown

  echo "- Fetching logs from testing pods in Running status due to running afer test suite timeout..."
  kyma test logs "${testsuiteName}" --test-status Running
  exit 1
fi
echo "  Generate junit results"
kyma test status "${testsuiteName}" -ojunit | sed 's/ (executions: [0-9]*)"/"/g' > junit_kyma_octopus-test-suite.xml
echo "--> Success"

