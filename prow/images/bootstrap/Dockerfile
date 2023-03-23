# Basic setup for Prow jobs

FROM debian:11

# Add commit details

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit


WORKDIR /workspace
RUN mkdir -p /workspace
ENV WORKSPACE=/workspace \
    TERM=xterm

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    ca-certificates \
    curl \
    git \
    jq \
    dnsutils \
    openssh-client \
    python2 \
    python-is-python2 \
    wget \
    apt-transport-https \
    gnupg2 \
    software-properties-common \
    lsb-release \
    gettext \
    libpcsclite1 \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install gcloud
ENV CLOUD_SDK_VERSION=410.0.0
ENV PATH=/google-cloud-sdk/bin:/workspace:${PATH} \
    CLOUDSDK_CORE_DISABLE_PROMPTS=1

RUN curl -fLSs -o gc-sdk.tar.gz https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${CLOUD_SDK_VERSION}-linux-x86_64.tar.gz && \
    tar xzf gc-sdk.tar.gz -C / && \
    rm gc-sdk.tar.gz && \
    /google-cloud-sdk/install.sh \
    --disable-installation-options \
    --bash-completion=false \
    --path-update=false \
    --usage-reporting=false && \
    gcloud components install alpha beta kubectl docker-credential-gcr && \
    gcloud components install gke-gcloud-auth-plugin &&\
    gcloud info | tee /workspace/gcloud-info.txt

ENV USE_GKE_GCLOUD_AUTH_PLUGIN=True
# Cluster Version
ENV CLUSTER_VERSION=1.25

# Kubectl 1.25
RUN mv /google-cloud-sdk/bin/kubectl.${CLUSTER_VERSION} /google-cloud-sdk/bin/kubectl

# Docker-in-docker
# hadolint ignore=SC1091
RUN . /etc/os-release && \
    curl -fsSL "https://download.docker.com/linux/$ID/gpg" \
    | apt-key add - && \
    add-apt-repository \
    "deb [arch=amd64] https://download.docker.com/linux/$ID $VERSION_CODENAME stable"

ENV DOCKER_VERSION=5:20.10.14*

RUN apt-get update && \
    apt-get install -y --no-install-recommends docker-ce=${DOCKER_VERSION} && \
    sed -i 's/cgroupfs_mount$/#cgroupfs_mount\n/' /etc/init.d/docker && \
    rm -rf /var/lib/apt/lists/*

# Move Docker's storage location
# hadolint ignore=SC2016
RUN echo 'DOCKER_OPTS="${DOCKER_OPTS} --data-root=/docker-graph"' | \
    tee --append /etc/default/docker
RUN mkdir /docker-graph

#spectral openapi validator
RUN curl -L https://raw.github.com/stoplightio/spectral/v6.0.0/scripts/install.sh | sh

# Prow Tools
# hadolint ignore=DL3022
COPY --from=eu.gcr.io/kyma-project/test-infra/prow-tools:v20221118-5c35cc55 /prow-tools /prow-tools
# for better access to prow-tools
ENV PATH=$PATH:/prow-tools

#################################################################
################ Install Cosign ####################
#################################################################
ENV COSIGN_VERSION=1.0.0
RUN curl -fLSs -o /usr/local/bin/cosign https://github.com/sigstore/cosign/releases/download/v${COSIGN_VERSION}/cosign-linux-amd64 &&\
    chmod +x /usr/local/bin/cosign

# The following lines needed for using dind since debian is using nftables by default
RUN update-alternatives --set iptables /usr/sbin/iptables-legacy && \
    update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy

RUN curl -fLSs -o shellcheck.tar.xz https://github.com/koalaman/shellcheck/releases/download/v0.5.0/shellcheck-v0.5.0.linux.x86_64.tar.xz \
    && tar xvf shellcheck.tar.xz \
    && mv shellcheck-v0.5.0/shellcheck /usr/bin/ \
    && rm -rf shellcheck-v0.5.0 shellcheck.tar.xz

RUN curl -fLSs -o /usr/local/bin/hadolint https://github.com/hadolint/hadolint/releases/download/v2.7.0/hadolint-Linux-x86_64 && \
    chmod +x /usr/local/bin/hadolint

#Install yq
ENV YQ_VERSION=v4.2.0 \
    YQ_BINARY=yq_linux_amd64
RUN curl -fLS -o /usr/bin/yq https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/${YQ_BINARY} && chmod +x /usr/bin/yq
