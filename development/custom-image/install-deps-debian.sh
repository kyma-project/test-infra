#!/bin/bash

###
# Following script installs necessary tooling for Debian to deploy Kyma on k3d.
#
# REQUIREMENTS:
# 64-bit version of one of these Debian versions:
# 
# - Stretch 9 (stable)
# - Jessie 8 (LTS)
# 
###

set -o errexit
set -o pipefail

MINIKUBE_VERSION=v1.28.0
KUBECTL_CLI_VERSION=v1.24.7
CRICTL_VERSION=v1.12.0
HELM_VERSION="v3.7.2"
DOCKER_VERSION=5:20.10.21~3-0~debian-bullseye
NODEJS_VERSION="14.x"
K3D_VERSION="5.0.1"
PG_MIGRATE_VERSION=v4.15.1
GO_VERSION=1.19.4

# install docker
sudo apt-get update
sudo apt-get upgrade -y
sudo apt-get install -y \
     apt-transport-https \
     ca-certificates \
     curl \
     gnupg2 \
     socat \
     lsb-release \
     wget \
     build-essential \
     conntrack \
     software-properties-common \
     postgresql-client-13 \
     pkg-config \
     libgit2-dev

curl -fsSL https://download.docker.com/linux/debian/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

echo \
  "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt update
sudo apt-cache policy docker-ce
sudo apt install -y docker-ce=${DOCKER_VERSION}

# install kubectl
curl -Lo /tmp/kubectl https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_CLI_VERSION}/bin/linux/amd64/kubectl && \
 chmod +x /tmp/kubectl && \
 sudo mv /tmp/kubectl /usr/local/bin/kubectl

# install helm
wget https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz -O - | tar -xzO linux-amd64/helm > /tmp/helm && \
	chmod +x /tmp/helm && \
	sudo mv /tmp/helm /usr/local/bin/helm && \
   rm -rf helm-${HELM_VERSION}-linux-amd64.tar.gz linux-amd64

# install minikube
curl -Lo /tmp/minikube https://storage.googleapis.com/minikube/releases/${MINIKUBE_VERSION}/minikube-linux-amd64 && \
chmod +x /tmp/minikube && \
sudo install /tmp/minikube /usr/local/bin/minikube


# install postgres and migrate tool
wget https://github.com/golang-migrate/migrate/releases/download/${PG_MIGRATE_VERSION}/migrate.linux-amd64.tar.gz -O - | tar -zxO migrate > /tmp/migrate && \
 chmod +x /tmp/migrate && \
 sudo mv /tmp/migrate /usr/local/bin/migrate


# install circtl
wget https://github.com/kubernetes-sigs/cri-tools/releases/download/${CRICTL_VERSION}/crictl-${CRICTL_VERSION}-linux-amd64.tar.gz
sudo tar zxvf crictl-${CRICTL_VERSION}-linux-amd64.tar.gz -C /usr/local/bin
rm -f crictl-${CRICTL_VERSION}-linux-amd64.tar.gz

# install jq and nodejs
curl -sL https://deb.nodesource.com/setup_${NODEJS_VERSION} | sudo bash -
sudo apt-get -y install \
     jq \
     nodejs

# install k3d
wget -q -O - https://raw.githubusercontent.com/rancher/k3d/main/install.sh | TAG=v${K3D_VERSION} bash

# install cloud-ops agent
# https://cloud.google.com/logging/docs/agent/installation
curl -sSO https://dl.google.com/cloudagents/add-google-cloud-ops-agent-repo.sh
sudo bash add-google-cloud-ops-agent-repo.sh \
  --also-install \
  --version=2.*.*

# install go
sudo mkdir /usr/local/go && \
     curl -fsSL -o /tmp/go.tar.gz "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" && \
     sudo tar xzf /tmp/go.tar.gz -C /usr/local && \
     rm /tmp/go.tar.gz
# shellcheck disable=SC2016
echo 'export PATH="$PATH:/usr/local/go/bin"' | sudo tee -a /etc/profile

# pre-fetch-docker-images
sudo docker pull eu.gcr.io/kyma-project/external/cypress/included:8.7.0
sudo docker pull eu.gcr.io/kyma-project/test-infra/docker-registry-2:20200202

sudo sed -i 's/\(GRUB_CMDLINE_LINUX_DEFAULT="\)\(.*\)\("\)/\1\2 systemd.legacy_systemd_cgroup_controller=false systemd.unified_cgroup_hierarchy=false\3/' /etc/default/grub
sudo sed -i 's/\(GRUB_CMDLINE_LINUX="\)\(.*\)\("\)/\1\2 systemd.legacy_systemd_cgroup_controller=false systemd.unified_cgroup_hierarchy=false\3/' /etc/default/grub
sudo update-grub
