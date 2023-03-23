# Basic setup for Prow jobs

FROM debian:11

# Commit details

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

#################################################################
###################### Basic Bootstrap ##########################
#################################################################
ARG K8S_VERSION=1.25
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
    python2 \
    python-is-python2 \
    wget \
    apt-transport-https \
    gnupg2 \
    shellcheck \
    software-properties-common \
    lsb-release \
    gettext \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Docker-in-docker
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
# hadolint ignore=SC1091
RUN . /etc/os-release && \
    curl -fsSL "https://download.docker.com/linux/$ID/gpg" | apt-key add - && \
    add-apt-repository \
    "deb [arch=amd64] https://download.docker.com/linux/$ID $VERSION_CODENAME stable"

ENV DOCKER_VERSION=5:20.10.14*

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    docker-ce=${DOCKER_VERSION} docker-ce-cli=${DOCKER_VERSION} containerd.io && \
    sed -i 's/cgroupfs_mount$/#cgroupfs_mount\n/' /etc/init.d/docker && \
    rm -rf /var/lib/apt/lists/*

# Move Docker's storage location
# hadolint ignore=SC2016
RUN echo 'DOCKER_OPTS="${DOCKER_OPTS} --data-root=/docker-graph"' | \
    tee --append /etc/default/docker && \
    mkdir /docker-graph

#Use iptables instead of nftables

RUN update-alternatives --set iptables /usr/sbin/iptables-legacy && \
    update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy

#################################################################
###################### GOLANG ###################################
#################################################################

# Versions

ENV GO_VERSION 1.19.4
ENV DEP_RELEASE_TAG v0.5.4

# Install Go
ENV GOPATH /workspace/go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN curl -fsSL -o go.tar.gz "https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz" && \
    tar xzf go.tar.gz && \
    rm go.tar.gz && \
    mv go /usr/local && \
    mkdir -p "${GOPATH}/bin" && \
    mkdir -p "${GOPATH}/src"

# Install Dep

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh


#################################################################
###################### Node.js ##################################
#################################################################

RUN curl -fLSs https://deb.nodesource.com/setup_14.x | bash -
RUN apt-get -y --no-install-recommends install nodejs

#################################################################
###################### Cluster Integration ######################
#################################################################

#Install helm

ENV HELM_VERSION="v3.7.2"

RUN curl -fLSs https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz | tar -xzO linux-amd64/helm > /usr/local/bin/helm && \
	  chmod +x /usr/local/bin/helm && \
      rm -rf helm-${HELM_VERSION}-linux-amd64.tar.gz linux-amd64

# Install Azure CLI
# If you receive error related to recv-keys please make sure you have latest signing key
# https://docs.microsoft.com/en-us/cli/azure/install-azure-cli-apt?view=azure-cli-latest#signingKey
# Install packages required by az

RUN apt-get update && apt-get install -y --no-install-recommends \
    apt-transport-https \
    lsb-release \
    software-properties-common \
    dirmngr \
    gnupg2 && \
    rm -rf /var/lib/apt/lists/*

ENV AZURE_CLI_VERSION="2.34.1-1*"

RUN curl -sL https://packages.microsoft.com/keys/microsoft.asc | gpg --dearmor | \
    tee /etc/apt/trusted.gpg.d/microsoft.asc.gpg > /dev/null && \
    echo "deb [arch=amd64] https://packages.microsoft.com/repos/azure-cli/ $(lsb_release -cs) main" | \
    tee /etc/apt/sources.list.d/azure-cli.list && \
    apt-get update && \
    apt-get install -y --no-install-recommends "azure-cli=${AZURE_CLI_VERSION}" && \
    rm -rf /var/lib/apt/lists/*

#################################################################
########################## Prow Tools ###########################
#################################################################

# hadolint ignore=DL3022
COPY --from=eu.gcr.io/kyma-project/test-infra/prow-tools:v20221118-5c35cc55 /prow-tools /prow-tools
# for better access to prow-tools
ENV PATH=$PATH:/prow-tools


#################################################################
######################## Install gcloud #########################
#################################################################

ENV CLOUD_SDK_VERSION=410.0.0 \
    PATH=/google-cloud-sdk/bin:/workspace:${PATH} \
    CLOUDSDK_CORE_DISABLE_PROMPTS=1 \
    CLUSTER_VERSION=$K8S_VERSION

RUN curl -fLSs -o gc-sdk.tar.gz https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${CLOUD_SDK_VERSION}-linux-x86_64.tar.gz && \
    tar xzf gc-sdk.tar.gz -C / && \
    rm gc-sdk.tar.gz && \
    /google-cloud-sdk/install.sh \
    --disable-installation-options \
    --bash-completion=false \
    --path-update=false \
    --usage-reporting=false && \
    gcloud components install alpha beta kubectl docker-credential-gcr && \
    gcloud components install gke-gcloud-auth-plugin && \
    gcloud info | tee /workspace/gcloud-info.txt && \
    mv /google-cloud-sdk/bin/kubectl.${CLUSTER_VERSION} /google-cloud-sdk/bin/kubectl

ENV USE_GKE_GCLOUD_AUTH_PLUGIN=True
#################################################################
################ Install Benchmark utilities ####################
#################################################################

RUN go install golang.org/x/perf/cmd/...

#################################################################
####################### Install smctl ###########################
#################################################################

ENV SMCTL_VERSION=v1.11.12
RUN curl -fLSs -o smctl.tar.gz https://github.com/Peripli/service-manager-cli/releases/download/${SMCTL_VERSION}/smctl-${SMCTL_VERSION}-linux-amd64.tar.gz && \
    tar xzf smctl.tar.gz && \
    rm smctl.tar.gz && \
    mv smctl /usr/local/go/bin

#################################################################
################ Install Cosign #################################
#################################################################
ENV COSIGN_VERSION=1.0.0
RUN curl -fLSs -o /usr/local/bin/cosign https://github.com/sigstore/cosign/releases/download/v${COSIGN_VERSION}/cosign-linux-amd64 &&\
    chmod +x /usr/local/bin/cosign

#################################################################
################ Install kcp-cli ################################
#################################################################

# Note: kcp-cli doesn't provide releases
ENV KCPCLI_VERSION=master-8a4162fd
RUN curl -fLSs -o /usr/local/bin/kcp https://storage.googleapis.com/kyma-development-artifacts/kcp/${KCPCLI_VERSION}/kcp-linux &&\
    chmod +x /usr/local/bin/kcp

#################################################################
################ Install yq (yaml tool) #########################
#################################################################
ENV YQ_VERSION=v4.15.1
RUN curl -fLSs -o /usr/local/bin/yq https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_linux_amd64 &&\
    chmod +x /usr/local/bin/yq

#################################################################
################ Install gh (github-cli) ########################
#################################################################
ENV GH_VERSION=2.2.0
RUN curl -fsSL https://github.com/cli/cli/releases/download/v${GH_VERSION}/gh_${GH_VERSION}_linux_amd64.tar.gz -o - | tar -xz &&\
  mv gh_${GH_VERSION}_linux_amd64/bin/gh /usr/local/bin/gh &&\
  chmod +x /usr/local/bin/gh &&\
  rm -r gh_${GH_VERSION}_linux_amd64

#################################################################
################ Install kustomize ########################
#################################################################
ENV KUSTOMIZE_VERSION=4.5.7
RUN curl -fsSL https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv${KUSTOMIZE_VERSION}/kustomize_v${KUSTOMIZE_VERSION}_linux_amd64.tar.gz -o - | tar -xz &&\
  mv kustomize /usr/local/bin/kustomize &&\
  chmod +x /usr/local/bin/kustomize
