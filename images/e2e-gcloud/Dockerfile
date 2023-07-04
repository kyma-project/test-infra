FROM golang:1.20.4-alpine as base

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
    python3 \
    openssh-client \
    py3-crcmod \
    py3-openssl \
    gnupg \
    libc6-compat \
    gettext

ARG DUMB_INIT_VERSION=1.2.5
RUN curl -Lo /usr/local/bin/dumb-init "https://github.com/Yelp/dumb-init/releases/download/v${DUMB_INIT_VERSION}/dumb-init_${DUMB_INIT_VERSION}_$(uname -m)" && \
    chmod +x /usr/local/bin/dumb-init

RUN adduser -D prow

ENV USE_GKE_GCLOUD_AUTH_PLUGIN=True
ENV PATH=/google-cloud-sdk/bin:${PATH}
ARG GCLOUD_CLI_VERSION=437.0.1

RUN curl -fLSs -o gc-sdk.tar.gz "https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-cli-${GCLOUD_CLI_VERSION}-linux-$(uname -m).tar.gz" && \
    tar xzf gc-sdk.tar.gz -C / && \
    rm gc-sdk.tar.gz && \
    gcloud components install alpha beta kubectl docker-credential-gcr gke-gcloud-auth-plugin && \
    gcloud config set core/disable_usage_reporting true && \
    gcloud config set component_manager/disable_update_check true && \
    gcloud config set metrics/environment github_docker_image && \
    gcloud --version; \
    kubectl version --client

ARG HELM_VERSION=v3.12.1
RUN curl -Lo helm.tar.gz "https://get.helm.sh/helm-${HELM_VERSION}-linux-$(go env GOARCH).tar.gz" && \
    tar -xzOf helm.tar.gz linux-amd64/helm > /usr/local/bin/helm && \
    chmod +x /usr/local/bin/helm

FROM base AS tools

RUN CGO_ENABLED=0 go install github.com/kyma-project/test-infra/development/jobguard/cmd/jobguard@main

FROM base

COPY --from=tools /go/bin/jobguard /usr/local/bin/jobguard


USER prow:prow

ENTRYPOINT ["/usr/local/bin/dumb-init", "--"]
