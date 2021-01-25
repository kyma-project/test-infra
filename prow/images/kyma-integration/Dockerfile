# Basic setup for Prow jobs

FROM alpine:3

# Commit details

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

#################################################################
###################### Basic Bootstrap ##########################
#################################################################
ARG K8S_VERSION=1.16
ARG commit
ENV IMAGE_COMMIT=$commit
LABEL IMAGE_COMMIT=$commit
WORKDIR /workspace
RUN mkdir -p /workspace
ENV WORKSPACE=/workspace \
    TERM=xterm

RUN apk update && apk add \
    bash \
    sudo \
    build-base \
    coreutils \
    ca-certificates \
    curl \
    git \
    jq \
    bind-tools \
    openssh-client \
    python3 \
    wget \
    gnupg \
    shellcheck \
    iptables \
    ip6tables \
    gettext \
    git \
    tar \ 
    dpkg \
    libc6-compat \
    openssh \
    openssl

# Docker-in-docker
# TODO semantic versioning
# ENV DOCKER_VERSION=5:19.03.8~3-0~debian-buster

# RUN apk add docker=${DOCKER_VERSION} docker-cli=${DOCKER_VERSION} containerd && \
RUN apk add docker docker-openrc containerd && \
    sed -i 's/cgroupfs_mount$/#cgroupfs_mount\n/' /etc/init.d/docker

# Move Docker's storage location

RUN mkdir -p /etc/default/ && \
    echo 'DOCKER_OPTS="${DOCKER_OPTS} --data-root=/docker-graph"' | \
    tee --append /etc/default/docker && \
    mkdir /docker-graph

#Use iptables instead of nftables
RUN update-alternatives --install /sbin/iptables iptables /sbin/iptables-legacy 10 && \
    update-alternatives --install /sbin/ip6tables ip6tables /sbin/ip6tables-legacy 10

#################################################################
###################### GOLANG ###################################
#################################################################

# Versions

ENV GO_VERSION 1.14.2
ENV DEP_RELEASE_TAG v0.5.4

# Install Go
ENV GOPATH /workspace/go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN wget -q https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz && \
    tar xzf go${GO_VERSION}.linux-amd64.tar.gz && \
    rm go${GO_VERSION}.linux-amd64.tar.gz && \
    mv go /usr/local && \
    mkdir -p ${GOPATH}/bin && \
    mkdir -p ${GOPATH}/src

# Install Dep

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh


#################################################################
###################### Node.js ##################################
#################################################################

# TODO proper version
# RUN curl -sL https://deb.nodesource.com/setup_14.x | sudo bash -
# RUN apt -y install nodejs
RUN apk add nodejs

#################################################################
###################### Cluster Integration ###################### 
#################################################################

#Install helm

ENV HELM_VERSION="v3.2.1"

RUN wget https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz -O - | tar -xzO linux-amd64/helm > /usr/local/bin/helm &&\
	  chmod +x /usr/local/bin/helm &&\
      rm -rf helm-${HELM_VERSION}-linux-amd64.tar.gz linux-amd64

# Build and install Azure CLI

ENV AZURE_CLI_VERSION="2.2.0"

RUN apk add \
    py3-pip \
    python3-dev \
    libffi-dev \
    openssl-dev && \
    pip install azure-cli==2.2.0

#################################################################
########################## Prow Tools ###########################
#################################################################

COPY --from=eu.gcr.io/kyma-project/test-infra/prow-tools:v20201113-72ed0661 /prow-tools /prow-tools
# for better access to prow-tools
ENV PATH=$PATH:/prow-tools


#################################################################
######################## Install gcloud #########################
#################################################################

ENV CLOUD_SDK_VERSION=314.0.0 \
    PATH=/google-cloud-sdk/bin:/workspace:${PATH} \
    CLOUDSDK_CORE_DISABLE_PROMPTS=1 \
    CLUSTER_VERSION=$K8S_VERSION

RUN wget -q https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${CLOUD_SDK_VERSION}-linux-x86_64.tar.gz && \
    tar xzf google-cloud-sdk-${CLOUD_SDK_VERSION}-linux-x86_64.tar.gz -C / && \
    rm google-cloud-sdk-${CLOUD_SDK_VERSION}-linux-x86_64.tar.gz && \
    /google-cloud-sdk/install.sh \
    --disable-installation-options \
    --bash-completion=false \
    --path-update=false \
    --usage-reporting=false && \
    gcloud components install alpha beta kubectl docker-credential-gcr && \
    gcloud info | tee /workspace/gcloud-info.txt && \
    mv /google-cloud-sdk/bin/kubectl.${CLUSTER_VERSION} /google-cloud-sdk/bin/kubectl
