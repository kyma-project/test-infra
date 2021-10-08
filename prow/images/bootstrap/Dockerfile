# Basic setup for Prow jobs

FROM debian:bullseye

# Add commit details

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit


WORKDIR /workspace
RUN mkdir -p /workspace
ENV WORKSPACE=/workspace \
    TERM=xterm

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    ca-certificates \
    curl \
    git \
    jq \
    dnsutils \
    openssh-client \
    python \
    wget \
    apt-transport-https \
    gnupg2 \
    software-properties-common \
    lsb-release \
    gettext \
    libpcsclite1 \
    && apt-get clean

# Install gcloud
ENV CLOUD_SDK_VERSION=358.0.0
ENV PATH=/google-cloud-sdk/bin:/workspace:${PATH} \
    CLOUDSDK_CORE_DISABLE_PROMPTS=1

RUN wget -q https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${CLOUD_SDK_VERSION}-linux-x86_64.tar.gz && \
    tar xzf google-cloud-sdk-${CLOUD_SDK_VERSION}-linux-x86_64.tar.gz -C / && \
    rm google-cloud-sdk-${CLOUD_SDK_VERSION}-linux-x86_64.tar.gz && \
    /google-cloud-sdk/install.sh \
    --disable-installation-options \
    --bash-completion=false \
    --path-update=false \
    --usage-reporting=false && \
    gcloud components install alpha beta kubectl docker-credential-gcr && \
    gcloud info | tee /workspace/gcloud-info.txt

# Cluster Version
ENV CLUSTER_VERSION=1.20

# Kubectl 1.20
RUN mv /google-cloud-sdk/bin/kubectl.${CLUSTER_VERSION} /google-cloud-sdk/bin/kubectl

# Docker-in-docker
RUN curl -fsSL https://download.docker.com/linux/$(. /etc/os-release; echo "$ID")/gpg \
    | apt-key add - && \
    add-apt-repository \
    "deb [arch=amd64] https://download.docker.com/linux/$(. /etc/os-release; echo "$ID") \
    $(lsb_release -cs) stable"

ENV DOCKER_VERSION=5:20.10.8*

RUN apt-get update && \
    apt-get install -y --no-install-recommends docker-ce=${DOCKER_VERSION} && \
    sed -i 's/cgroupfs_mount$/#cgroupfs_mount\n/' /etc/init.d/docker

# Move Docker's storage location
RUN echo 'DOCKER_OPTS="${DOCKER_OPTS} --data-root=/docker-graph"' | \
    tee --append /etc/default/docker
RUN mkdir /docker-graph

#spectral openapi validator
RUN curl -L https://raw.github.com/stoplightio/spectral/v6.0.0/scripts/install.sh | sh

# Prow Tools

COPY --from=eu.gcr.io/kyma-project/test-infra/prow-tools:v20210401-294e46e5 /prow-tools /prow-tools
# for better access to prow-tools
ENV PATH=$PATH:/prow-tools

#################################################################
################ Install Cosign ####################
#################################################################
ENV COSIGN_VERSION=1.0.0
RUN wget https://github.com/sigstore/cosign/releases/download/v${COSIGN_VERSION}/cosign-linux-amd64 -O /usr/local/bin/cosign &&\
    chmod +x /usr/local/bin/cosign

# The following lines needed for using dind since debian is using nftables by default
RUN update-alternatives --set iptables /usr/sbin/iptables-legacy && \
    update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy
ENV SHELLCHECK_VERSION=0.4.6

RUN wget https://github.com/koalaman/shellcheck/releases/download/v${SHELLCHECK_VERSION}/shellcheck-v${SHELLCHECK_VERSION}.linux.x86_64.tar.xz \
    && tar xvf shellcheck-v${SHELLCHECK_VERSION}.linux.x86_64.tar.xz \
    && mv shellcheck-v${SHELLCHECK_VERSION}/shellcheck /usr/bin/ \
    && rm -rf shellcheck-v${SHELLCHECK_VERSION} shellcheck-v${SHELLCHECK_VERSION}.linux.x86_64.tar.xz

ADD .shellcheckrc /root/.shellcheckrc
