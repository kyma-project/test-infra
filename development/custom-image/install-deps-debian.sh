#!/bin/bash

###
# Following script installs necessary tooling for Debian to deploy Kyma on Minikube.
#
# REQUIREMENTS:
# 64-bit version of one of these Debian versions:
# 
# - Stretch 9 (stable)
# - Jessie 8 (LTS)
# 
###

set -o errexit

MINIKUBE_VERSION=v1.14.2
KUBECTL_CLI_VERSION=v1.19.7
CRICTL_VERSION=v1.12.0
HELM_VERSION="v3.4.2"
DOCKER_VERSION=5:20.10.5~3-0~debian-buster
NODEJS_VERSION="14.x"

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
     software-properties-common

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
 sudo mv /tmp/minikube /usr/local/bin/minikube

# install circtl
wget https://github.com/kubernetes-sigs/cri-tools/releases/download/${CRICTL_VERSION}/crictl-${CRICTL_VERSION}-linux-amd64.tar.gz
sudo tar zxvf crictl-${CRICTL_VERSION}-linux-amd64.tar.gz -C /usr/local/bin
rm -f crictl-${CRICTL_VERSION}-linux-amd64.tar.gz

# install jq and nodejs
curl -sL https://deb.nodesource.com/setup_${NODEJS_VERSION} | sudo bash -
sudo apt-get -y install \
     jq \
     nodejs

# install monitoring agent
# https://cloud.google.com/monitoring/agent/installation
curl -sSO https://dl.google.com/cloudagents/add-monitoring-agent-repo.sh && \
sudo bash add-monitoring-agent-repo.sh && \
sudo apt-get update
sudo apt-cache madison stackdriver-agent
sudo apt-get install -y 'stackdriver-agent=6.*'

# install logging agent
# https://cloud.google.com/logging/docs/agent/installation
curl -sSO https://dl.google.com/cloudagents/add-logging-agent-repo.sh && \
sudo bash add-logging-agent-repo.sh && \
sudo apt-get update
sudo apt-cache madison google-fluentd
sudo apt-get install -y 'google-fluentd=1.*'
sudo apt-get install -y google-fluentd-catch-all-config

# pre-fetch-docker-images
sudo docker pull eu.gcr.io/kyma-project/external/cypress/included:7.5.0
sudo docker pull eu.gcr.io/kyma-project/test-infra/docker-registry-2:20200202

