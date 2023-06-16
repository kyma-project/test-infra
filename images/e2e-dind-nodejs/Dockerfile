FROM node:16.20.0-bookworm-slim as base


ENV TERM=xterm
ENV DEBIAN_FRONTEND=noninteractive


SHELL ["/bin/bash", "-o", "pipefail", "-c"]
#TODO Pin chrome and Docker to specific version if needed
#TODO adjust required packages with --no-install-recommends
# hadolint ignore=DL3015
RUN set -eux; \
    apt-get update && apt-get install -y \
    curl \
    iptables \
    build-essential \
    libgtk2.0-0 libgtk-3-0 libgbm-dev libnotify-dev libgconf-2-4 libnss3 libxss1 libasound2 libxtst6 xauth xvfb && \
    curl -sL https://get.docker.com | sh - && \
    update-alternatives --set iptables /usr/sbin/iptables-legacy && \
    update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy && \
    curl -sL https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add - && \
    sh -c 'echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" > /etc/apt/sources.list.d/google-chrome.list' && \
    apt-get update && apt-get install -y google-chrome-stable && \
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

FROM golang:1.20.4 AS go-deps
ENV CGO_ENABLED=0
ARG K3D_VERSION=v5.4.9

RUN apt-get install --no-install-recommends -y curl

RUN curl -Lso install.sh https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh && \
    chmod +x ./install.sh &&\
    TAG=${K3D_VERSION} ./install.sh &&\
    rm install.sh

# Only k3d is needed atm. If it's proven unstable, switch to kind
#ARG KIND_VERSION=v0.18.0
#RUN go install sigs.k8s.io/kind@${KIND_VERSION}

ARG CLUSTER_VERSION=v1.26.3
RUN curl -Lo /usr/local/bin/kubectl "https://storage.googleapis.com/kubernetes-release/release/${CLUSTER_VERSION}/bin/linux/$(go env GOARCH)/kubectl" && \
    chmod +x /usr/local/bin/kubectl

FROM base

COPY init.sh /
COPY --from=go-deps /usr/local/bin/k3d /usr/local/bin/
COPY --from=go-deps /usr/local/bin/kubectl /usr/local/bin/

ENTRYPOINT [ "/init.sh" ]