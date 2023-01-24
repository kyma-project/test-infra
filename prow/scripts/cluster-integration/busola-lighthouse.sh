#!/usr/bin/env bash

set -o errexit
set -o pipefail

LOCAL_KYMA_DIR="./local-kyma"
K3S_DOMAIN="local.kyma.dev"
PLAYWRIGHT_IMAGE="mcr.microsoft.com/playwright:v1.15.0-focal"

function install_cli() {
  local install_dir
  declare -r install_dir="/usr/local/bin"
  mkdir -p "$install_dir"

  local os
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  if [[ -z "$os" || ! "$os" =~ ^(darwin|linux)$ ]]; then
    echo >&2 -e "Unsupported host OS. Must be Linux or Mac OS X."
    exit 1
  else
    readonly os
  fi

  pushd "$install_dir" || exit
  curl -Lo kyma "https://storage.googleapis.com/kyma-cli-stable/kyma-${os}"
  chmod +x kyma
  popd

  kyma version --client
}

generate_cert(){
    # $1 is the domain
    mkdir ssl
    pushd ssl
    
    # Generate private key
    openssl genrsa -out ca.key 2048
    
    # Generate root certificate
    openssl req -x509 -new -nodes -subj "/C=US/O=_Development CA/CN=Development certificates" -key ca.key -sha256 -days 3650 -out ca.crt
    
    echo "Root certificate generated √"
    
    # Generate a private key
    openssl genrsa -out "$1.key" 2048
    
    # Create a certificate signing request
    openssl req -new -subj "/C=US/O=Local Development/CN=$1" -key "$1.key" -out "$1.csr"
    
    # Create a config file for the extensions
>"$1.ext" cat <<-EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth, clientAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = $1
DNS.2 = *.$1
EOF
    
    # Create the signed certificate
    openssl x509 -req \
    -in "$1.csr" \
    -extfile "$1.ext" \
    -CA ca.crt \
    -CAkey ca.key \
    -CAcreateserial \
    -out "$1.crt" \
    -days 365 \
    -sha256
    
    echo "Certificate generated for the $1 domain √"
    popd
}

install_busola(){
    # $1 is the domain
    echo "Deploying Busola resources on the $1 domain"
    
    kubectl create secret tls default-ssl-certificate \
    --namespace kube-system \
    --key ./ssl/"${1}".key \
    --cert ./ssl/"${1}".crt
    
    helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
    # helm repo update
    
    helm install ingress-nginx --version 4.1.3 \
    --namespace=kube-system \
    --set controller.extraArgs.default-ssl-certificate=kube-system/default-ssl-certificate \
    ingress-nginx/ingress-nginx > /dev/null
    
    #wait for ingress controller to start
    kubectl wait --namespace kube-system \
    --for=condition=ready pod \
    --selector=app.kubernetes.io/component=controller \
    --timeout=120s
    
    pushd busola-resources
    
    for i in ./**{/*,}.yaml; do
        sed -i "s,%DOMAIN%,$1,g" "$i"
    done
    
    kubectl apply -k .
    
    popd
    echo "Busola resources applied √"
}

echo "Node.js version: $(node -v)"
echo "NPM version: $(npm -v)"

echo "STEP: Installing Kyma CLI fore easier cluster setup"
install_cli

echo "STEP: Preparing k3s cluster"
kyma provision k3d --ci

echo "STEP: Generating certificate"
generate_cert $K3S_DOMAIN

echo "STEP: Installing Busola on the cluster"
install_busola $K3S_DOMAIN



# wait for all Busola pods to be ready
kubectl wait \
--for=condition=ready pod \
--all \
--timeout=120s

mkdir -p "$PWD/busola-tests/fixtures"
mkdir -p "$PWD/busola-tests/test-results"
cp "$PWD/kubeconfig-kyma.yaml" "$PWD/busola-tests/fixtures/kubeconfig.yaml"

echo "STEP: Running Lighthouse audit inside Docker"
docker run --entrypoint /bin/bash --network=host -v "$PWD/busola-tests:/tests" -w /tests $PLAYWRIGHT_IMAGE -c "npm ci --no-optional; npx playwright install;  NO_COLOR=1 npx playwright test"

