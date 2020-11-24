#!/usr/bin/env bash

set -o errexit
set -o pipefail  # Fail a pipe if any sub-command fails.

date

export KANIKO_IMAGE="eu.gcr.io/kyma-project/external/aerfio/kaniko-executor:v1.3.0"
export DOMAIN=${KYMA_DOMAIN:-local.kyma.dev}
if [[ -z $REGISTRY_VALUES ]]; then
  export REGISTRY_VALUES="dockerRegistry.enableInternal=false,dockerRegistry.serverAddress=registry.localhost:5000,dockerRegistry.registryAddress=registry.localhost:5000,docker-registry.destinationRule.enabled=false,containers.manager.envs.functionBuildExecutorImage.value=${KANIKO_IMAGE},images.manager.repository=aerfio/function-controller,images.manager.tag=latest"
fi

export KYMA_SOURCES_DIR="./kyma"

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

host::update_etc_hosts(){
  # needed for external docker registry
  echo "${REGISTRY_IP} registry.localhost" >> /etc/hosts
}

host::create_docker_registry(){
cat > registries.yaml <<EOL
mirrors:
  registry.localhost:5000:
    endpoint:
    - http://registry.localhost:5000
  configs: {}
  auths: {}
  
EOL

mkdir -p /etc/rancher/k3s
cp registries.yaml /etc/rancher/k3s

docker run -d \
  -p 5000:5000 \
  --restart=always \
  --name registry.localhost \
  -v "$PWD/registry:/var/lib/registry" \
  registry:2
}

install::prereq(){
    curl https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash
    apt-get -y install jq
}

install::k3s() {
    echo "--> Installing k3s"
    curl -sfL https://get.k3s.io | K3S_KUBECONFIG_MODE=777 INSTALL_K3S_EXEC="server --disable traefik" sh -
    mkdir -p ~/.kube
    cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
    chmod 600 ~/.kube/config
    k3s --version
    date
}

function host::patch_coredns() {
  echo "Patching CoreDns with REGISTRY_IP=$REGISTRY_IP"
  sed "s/REGISTRY_IP/$REGISTRY_IP/" coredns-patch.tpl >coredns-patch.yaml
  kubectl -n kube-system patch cm coredns --patch "$(cat coredns-patch.yaml)"
}

host::create_coredns_template
host::create_docker_registry
# shellcheck disable=SC2155
export REGISTRY_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' /registry.localhost)

host::update_etc_hosts

install::prereq

date
echo "--> Creating k8s cluster via k3s"
install::k3s

# shellcheck disable=SC2004
while [[ $(kubectl get nodes -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do echo "Waiting for cluster nodes to be ready, elapsed time: $(( $SECONDS/60 )) min $(( $SECONDS % 60 )) sec"; sleep 2; done
host::patch_coredns

kubectl apply -f "$KYMA_SOURCES_DIR/resources/cluster-essentials/files" -n kyma-system
helm upgrade --atomic --create-namespace -i serverless "$KYMA_SOURCES_DIR/resources/serverless" -n kyma-system --set "$REGISTRY_VALUES",global.ingress.domainName="$DOMAIN" --wait

echo "##############################################################################"
# shellcheck disable=SC2004
echo "# Serverless installed in $(( $SECONDS/60 )) min $(( $SECONDS % 60 )) sec"
echo "##############################################################################"

kubectl apply -f "$KYMA_SOURCES_DIR/components/function-controller/config/samples/serverless_v1alpha1_function.yaml"
echo "wait 180s for function to be ready"
kubectl wait --for=condition=Running function/demo --timeout 180s
echo "success!"
kubectl get -f "$KYMA_SOURCES_DIR/components/function-controller/config/samples/serverless_v1alpha1_function.yaml" -oyaml

exit 0
