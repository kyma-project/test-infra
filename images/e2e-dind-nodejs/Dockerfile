FROM node:20.13.1-bookworm-slim as base


ENV TERM=xterm
ENV DEBIAN_FRONTEND=noninteractive


SHELL ["/bin/bash", "-o", "pipefail", "-c"]
#TODO Pin chromium and Docker to specific version if needed
#TODO adjust required packages with --no-install-recommends
# hadolint ignore=DL3015
RUN set -eux; \
    apt-get update && apt-get upgrade -y && apt-get install -y \
    curl \
    iptables \
    build-essential \
    jq \
    chromium \
    libgtk2.0-0 libgtk-3-0 libgbm-dev libnotify-dev libgconf-2-4 libnss3 libxss1 libasound2 libxtst6 xauth xvfb && \
    curl -sL https://get.docker.com | sh - && \
    update-alternatives --set iptables /usr/sbin/iptables-legacy && \
    update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy && \
    rm -rf /var/cache/apt/* && apt-get clean

ARG DUMB_INIT_VERSION=1.2.5
RUN curl -Lo /usr/local/bin/dumb-init "https://github.com/Yelp/dumb-init/releases/download/v${DUMB_INIT_VERSION}/dumb-init_${DUMB_INIT_VERSION}_$(uname -m)" && \
    chmod +x /usr/local/bin/dumb-init

RUN set -eux; \
    addgroup -S dockremap 	&& \
    adduser -SDHs /sbin/nologin dockremap; \
    echo 'dockremap:165536:65536' >> /etc/subuid && \
    echo 'dockremap:165536:65536' >> /etc/subgid

RUN mkdir -p /workspace; \
    chmod 1777 /workspace

WORKDIR /workspace

FROM golang:1.22.3 AS go-deps
ENV CGO_ENABLED=0
ARG K3D_VERSION=v5.6.3

RUN apt-get install --no-install-recommends -y curl

RUN curl -Lso install.sh https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh && \
    chmod +x ./install.sh &&\
    TAG=${K3D_VERSION} ./install.sh &&\
    rm install.sh

ARG CLUSTER_VERSION=v1.28.10
RUN curl -Lo /usr/local/bin/kubectl "https://storage.googleapis.com/kubernetes-release/release/${CLUSTER_VERSION}/bin/linux/$(go env GOARCH)/kubectl" && \
    chmod +x /usr/local/bin/kubectl

ARG HELM_VERSION=v3.15.0
RUN curl -Lo helm.tar.gz "https://get.helm.sh/helm-${HELM_VERSION}-linux-$(go env GOARCH).tar.gz" && \
    tar -xzOf helm.tar.gz linux-amd64/helm > /usr/local/bin/helm && \
    chmod +x /usr/local/bin/helm

FROM base

COPY init.sh /
COPY --from=go-deps /usr/local/bin/k3d /usr/local/bin/
COPY --from=go-deps /usr/local/bin/kubectl /usr/local/bin/
COPY --from=go-deps /usr/local/bin/helm /usr/local/bin/

ENTRYPOINT [ "/init.sh" ]
