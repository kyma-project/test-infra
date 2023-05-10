FROM golang:1.20.3-alpine AS base

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
    gettext

ARG DUMB_INIT_VERSION=1.2.5
RUN curl -Lo /usr/local/bin/dumb-init "https://github.com/Yelp/dumb-init/releases/download/v${DUMB_INIT_VERSION}/dumb-init_${DUMB_INIT_VERSION}_$(uname -m)" && \
    chmod +x /usr/local/bin/dumb-init

RUN adduser -D prow

ARG CLUSTER_VERSION=v1.25.9
RUN curl -Lo /usr/local/bin/kubectl "https://storage.googleapis.com/kubernetes-release/release/${CLUSTER_VERSION}/bin/linux/$(go env GOARCH)/kubectl" && \
    chmod +x /usr/local/bin/kubectl

FROM base AS tools

RUN CGO_ENABLED=0 go install github.com/kyma-project/test-infra/development/jobguard/cmd/jobguard@main

FROM base

COPY --from=tools /go/bin/jobguard /usr/local/bin/jobguard

USER prow:prow

ENTRYPOINT ["/usr/local/bin/dumb-init", "--"]