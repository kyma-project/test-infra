FROM golang:1.24.3-alpine3.21 as base

RUN set -eux; \
    apk upgrade --no-cache; \
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
    gettext \
    libgit2-dev \
    pkgconf

RUN adduser -D prow

ENV USE_GKE_GCLOUD_AUTH_PLUGIN=True
ENV PATH=/google-cloud-sdk/bin:${PATH}
ARG GCLOUD_CLI_VERSION=476.0.0

RUN curl -fLSs -o gc-sdk.tar.gz "https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-cli-${GCLOUD_CLI_VERSION}-linux-$(uname -m).tar.gz" && \
    tar xzf gc-sdk.tar.gz -C / && \
    rm gc-sdk.tar.gz && \
    # Users should use gcloud auth configure-docker to authenticate with GCR instead of docker-credential-gcr, see https://github.com/GoogleCloudPlatform/docker-credential-gcr?tab=readme-ov-file#introduction
    # gcloud components install alpha beta kubectl docker-credential-gcr gke-gcloud-auth-plugin && \
    gcloud components install alpha beta kubectl gke-gcloud-auth-plugin && \
    gcloud config set core/disable_usage_reporting true && \
    gcloud config set component_manager/disable_update_check true && \
    gcloud config set metrics/environment github_docker_image && \
    gcloud --version; \
    kubectl version --client; \
    rm -rf /google-cloud-sdk/.install

ARG HELM_VERSION=v3.15.0
RUN curl -Lo helm.tar.gz "https://get.helm.sh/helm-${HELM_VERSION}-linux-$(go env GOARCH).tar.gz" && \
    tar -xzOf helm.tar.gz "linux-$(go env GOARCH)/helm" > /usr/local/bin/helm && \
    chmod +x /usr/local/bin/helm

ENTRYPOINT ["/usr/local/bin/dumb-init", "--"]