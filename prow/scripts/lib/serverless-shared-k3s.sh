#!/usr/bin/env bash

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

collect_results(){
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
    echo "########################################################"
    kubectl logs -n kyma-system -l app=serverless --tail=-1


    echo "########################################################"
    echo "kubectl logs -n kyma-system -l app=serverless-webhook --tail=-1"
    echo "########################################################"
    kubectl logs -n kyma-system -l app=serverless-webhook --tail=-1


    echo "########################################################"
    echo "Get logs from all serverless jobs"
    echo "########################################################"
    ALL_TEST_NAMESPACES=$(kubectl get namespace --selector created-by=serverless-controller-manager-test   --no-headers -o custom-columns=name:.metadata.name)
    # shellcheck disable=SC2206
    ALL=($ALL_TEST_NAMESPACES)
    for NAMESPACE in "${ALL[@]}"
    do
      kubectl logs --namespace "${NAMESPACE}" --all-containers  --selector job-name --ignore-errors --prefix=true
    done
echo ""
}