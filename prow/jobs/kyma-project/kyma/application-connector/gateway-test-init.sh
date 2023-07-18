#!/bin/bash
set -e
k3d registry create k3d-registry --port 5000
k3d cluster create k3d --registry-use k3d-k3d-registry:5000
kubectl cluster-info
apk add openssl
curl -Lo kyma.tar.gz "https://github.com/kyma-project/cli/releases/download/$(curl -s https://api.github.com/repos/kyma-project/cli/releases/latest | grep tag_name | cut -d '"' -f 4)/kyma_Linux_x86_64.tar.gz" \
&& mkdir kyma-release && tar -C kyma-release -zxvf kyma.tar.gz && chmod +x kyma-release/kyma && mv kyma-release/kyma /usr/local/bin \
&& rm -rf kyma-release kyma.tar.gz && cd tests/components/application-connector && kyma deploy --ci --components-file ./resources/installation-config/mini-kyma-os.yaml --source=local --workspace "/home/prow/go/src/github.com/kyma-project/kyma" --verbose
make test -f Makefile.test-application-gateway
k3d cluster delete