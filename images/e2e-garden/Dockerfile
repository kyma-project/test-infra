FROM golang:1.22.2-alpine AS base

RUN set -eux; \
    apk add --no-cache \
    build-base \
    ca-certificates \
    curl \
    wget \
    bash \
    jq \
    yq \
    xfsprogs \
    xz \
    git \
    device-mapper \
    openssh-client \
    gnupg \
    libc6-compat \
    gettext \
    dumb-init \
    pkgconf

RUN adduser -D prow

ARG CLUSTER_VERSION=v1.28.6
RUN curl -Lo /usr/local/bin/kubectl "https://storage.googleapis.com/kubernetes-release/release/${CLUSTER_VERSION}/bin/linux/$(go env GOARCH)/kubectl" && \
    chmod +x /usr/local/bin/kubectl

ARG HELM_VERSION=v3.13.2
RUN curl -Lo helm.tar.gz "https://get.helm.sh/helm-${HELM_VERSION}-linux-$(go env GOARCH).tar.gz" && \
    tar -xzOf helm.tar.gz linux-amd64/helm > /usr/local/bin/helm && \
    chmod +x /usr/local/bin/helm
FROM base AS tools

RUN CGO_ENABLED=0 go install github.com/kyma-project/test-infra/cmd/jobguard@main

FROM base

COPY --from=tools /go/bin/jobguard /usr/local/bin/jobguard

#USER prow:prow

ENTRYPOINT ["/usr/local/bin/dumb-init", "--"]
