#!/usr/bin/env bash

set -o errexit
set -o pipefail

export LOCAL_KYMA_DIR="./local-kyma"

prepare_k3s() {
    pushd ${LOCAL_KYMA_DIR}
    ./create-cluster-k3s.sh
    popd
}

generate_cert(){
    # $1 is the domain
    mkdir ssl
    pushd ssl
    
    # Generate private key
    openssl genrsa -out ca.key 2048
    
    # Generate root certificate
    openssl req -x509 -new -nodes -subj "/C=US/O=_Development CA/CN=Development certificates" -key ca.key -sha256 -days 3650 -out ca.crt
    
    echo "Root certificate generated"
    
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
    
    echo "Certificate generated successfully"
    popd
}

install_busola(){
    
    
    kubectl create secret tls default-ssl-certificate \
    --namespace kube-system \
    --key ./ssl/${1}.key \
    --cert ./ssl/${1}.crt
    
    helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
    helm repo update
    
    helm install ingress-nginx \
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
        sed -i "s,%DOMAIN%,$1,g" $i
    done
    
    kubectl apply -k .
    
    echo "Busola Url:"
    echo "https://busola.$1"
    
    popd
}




prepare_k3s
generate_cert "local.kyma.dev"
install_busola "local.kyma.dev"

node -v
npm -v


kubectl cluster-info
kubectl config view
#K3S_SERVER=$(kubectl config view --minify -o jsonpath='{.clusters[].cluster.server}')
K3S_SERVER="local.kyma.dev"

echo "Deploying Busola resources on the ${K3S_SERVER} server"

docker pull cypress/included:7.2.0

kubectl get pods

cp $PWD/kubeconfig-kyma.yaml $PWD/busola-tests/fixtures/kubeconfig.yaml

echo "Running Cypress tests inside Docker..."
docker run --entrypoint /bin/bash --network=host -v $PWD/busola-tests:/tests -w /tests cypress/included:7.2.0 -c "npm ci; cypress run --browser chrome --headless"



# export KYMA_KUBECONFIG_PATH="${PWD}/kubeconfig-kyma.yaml"
# echo "KUBECONFIG: ${KYMA_KUBECONFIG_PATH}"

# export KUBECONFIG="${KYMA_KUBECONFIG_PATH}"

