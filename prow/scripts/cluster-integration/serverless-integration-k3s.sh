#!/usr/bin/env bash

set -o errexit
set -o pipefail  # Fail a pipe if any sub-command fails.

date

export KANIKO_IMAGE="eu.gcr.io/kyma-project/tpi/kaniko-executor:1.6.0-7b6e8d78"
export DOMAIN=${KYMA_DOMAIN:-local.kyma.dev}
if [[ -z $REGISTRY_VALUES ]]; then
  export REGISTRY_VALUES="dockerRegistry.enableInternal=false,dockerRegistry.serverAddress=registry.localhost:5000,dockerRegistry.registryAddress=registry.localhost:5000,containers.manager.envs.functionBuildExecutorImage.value=${KANIKO_IMAGE}"
fi

export USE_ALPHA=${USE_ALPHA:-false}

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

docker ps --format="{{.Names}}"| grep -q registry.localhost || \
docker run -d \
  -p 5000:5000 \
  --restart=always \
  --name registry.localhost \
  -v "$PWD/registry:/var/lib/registry" \
  eu.gcr.io/kyma-project/test-infra/docker-registry-2:20200202
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

install::k3s() {
    echo "--> Installing k3s"
    curl -sfL https://get.k3s.io | K3S_KUBECONFIG_MODE=777 INSTALL_K3S_VERSION="v1.19.7+k3s1" INSTALL_K3S_EXEC="server --disable traefik" sh -
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

date

if [ "$USE_ALPHA" == "true" ]; then
  echo "--> Installing kyma-cli"
  install::kyma_cli

  echo "--> Provisioning k3d cluster for Kyma"
  kyma provision k3d --ci

  echo "--> Deploying Serverless"
  # The python38 function requires 40M+ of memory to work. Mostly used by kubeless. I need to overrride the defaultPreset to M to avoid OOMkill.
  kyma deploy -p evaluation --component cluster-essentials,serverless --ci --value "$REGISTRY_VALUES" --value global.ingress.domainName="$DOMAIN" --value "serverless.webhook.values.function.resources.defaultPreset=M" -s local -w $KYMA_SOURCES_DIR

else
  host::create_coredns_template
  host::create_docker_registry
  # shellcheck disable=SC2155
  export REGISTRY_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' /registry.localhost)

  host::update_etc_hosts

  echo "--> Creating k8s cluster via k3s"
  install::k3s

  # shellcheck disable=SC2004
  while [[ $(kubectl get nodes -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do echo "Waiting for cluster nodes to be ready, elapsed time: $(( $SECONDS/60 )) min $(( $SECONDS % 60 )) sec"; sleep 2; done
  host::patch_coredns

  echo "--> Deploying cluster-essentials"
  kubectl apply -f "$KYMA_SOURCES_DIR/resources/cluster-essentials/files" -n kyma-system

  echo "--> Deploying Serverless"
  helm upgrade --atomic --create-namespace -i serverless "$KYMA_SOURCES_DIR/resources/serverless" -n kyma-system --set "$REGISTRY_VALUES",global.ingress.domainName="$DOMAIN" --wait
fi 

echo "##############################################################################"
# shellcheck disable=SC2004
echo "# Serverless installed in $(( $SECONDS/60 )) min $(( $SECONDS % 60 )) sec"
echo "##############################################################################"

# I know it's bad practice and kinda smelly to do this, but we have two nasty dataraces that might happen, and simple sleep solves them both:
# webhook might not be ready in time (but somehow it still accepts the function, we have an issue for that)
# runtime configmaps might now have been copied to that namespace, but it should be handled by https://github.com/kyma-project/kyma/pull/10026
########
sleep 60
########

SERVERLESS_CHART_DIR="${KYMA_SOURCES_DIR}/resources/serverless"
job_name="k3s-serverless-test"


helm install serverless-test "${SERVERLESS_CHART_DIR}/charts/k3s-tests" -n default -f "${SERVERLESS_CHART_DIR}/values.yaml" --set jobName="${job_name}"

job_status=""

# helm does not wait for jobs to complete even with --wait
# TODO but helm@v3.5 has a flag that enables that, get rid of this function once we use helm@v3.5
getjobstatus(){
while true; do
    echo "Test job not completed yet..."
    [[ $(kubectl get jobs $job_name -o jsonpath='{.status.conditions[?(@.type=="Failed")].status}') == "True" ]] && job_status=1 && echo "Test job failed" && break
    [[ $(kubectl get jobs $job_name -o jsonpath='{.status.conditions[?(@.type=="Complete")].status}') == "True" ]] && job_status=0 && echo "Test job completed successfully" && break
    sleep 5
done
}

getjobstatus

echo "####################"
echo "kubectl get pods -A"
echo "###################"
kubectl get pods -A
echo "########################"
echo "kubectl get functions -A"
echo "########################"
kubectl get functions -A
echo "########################################################"
echo "kubectl logs -n kyma-system -l app=serverless --tail=-1"
kubectl logs -n kyma-system -l app=serverless --tail=-1
echo "##############################################"
echo "kubectl logs -l job-name=${job_name} --tail=-1"
kubectl logs -l job-name=${job_name} --tail=-1
echo "###############"
echo ""

echo "Exit code ${job_status}"

exit $job_status
