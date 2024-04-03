FROM golang:1.22.1-alpine as base


RUN set -eux; \
    apk upgrade --no-cache && \
    apk add --no-cache \
    build-base \
    ca-certificates \
    curl \
    bash \
    jq \
    yq \
    docker \
    btrfs-progs \
    e2fsprogs \
    e2fsprogs-extra \
    iptables \
    xfsprogs \
    xz \
    git \
    fuse-overlayfs \
    device-mapper \
    openssh-client \
    libgit2-dev \
    pkgconf \
    gettext \
    dumb-init

RUN set -eux; \
    addgroup -S dockremap 	&& \
    adduser -SDHs /sbin/nologin dockremap; \
    echo 'dockremap:165536:65536' >> /etc/subuid && \
    echo 'dockremap:165536:65536' >> /etc/subgid

RUN mkdir -p /workspace; \
    chmod 1777 /workspace

WORKDIR /workspace

FROM base AS go-deps
ENV CGO_ENABLED=0
ARG K3D_VERSION=v5.6.0

RUN curl -Lso install.sh https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh && \
    chmod +x ./install.sh &&\
    TAG=${K3D_VERSION} ./install.sh &&\
    rm install.sh

ARG KIND_VERSION=v0.20.0
RUN go install sigs.k8s.io/kind@${KIND_VERSION}

ARG HELM_VERSION=v3.13.2
RUN curl -Lo helm.tar.gz "https://get.helm.sh/helm-${HELM_VERSION}-linux-$(go env GOARCH).tar.gz" && \
    tar -xzOf helm.tar.gz "linux-$(go env GOARCH)/helm" > /usr/local/bin/helm && \
    chmod +x /usr/local/bin/helm

ARG CLUSTER_VERSION=v1.28.6
RUN curl -Lo /usr/local/bin/kubectl "https://storage.googleapis.com/kubernetes-release/release/${CLUSTER_VERSION}/bin/linux/$(go env GOARCH)/kubectl" && \
    chmod +x /usr/local/bin/kubectl

FROM base AS tools

RUN CGO_ENABLED=0 go install github.com/kyma-project/test-infra/cmd/jobguard@main

FROM base

COPY init.sh /
COPY --from=go-deps /usr/local/bin/k3d /usr/local/bin/
COPY --from=go-deps /usr/local/bin/helm /usr/local/bin/
COPY --from=go-deps /go/bin/kind /usr/local/bin/
COPY --from=go-deps /usr/local/bin/kubectl /usr/local/bin/

COPY --from=tools /go/bin/jobguard /usr/local/bin/jobguard

ENTRYPOINT [ "/init.sh" ]
