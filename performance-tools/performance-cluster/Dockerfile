FROM debian:stretch

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
  shellcheck \
  software-properties-common \
  lsb-release \
  gettext-base \
  && apt-get clean

# Install gcloud
ENV CLOUD_SDK_VERSION=299.0.0
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

# Docker-in-docker
RUN curl -fsSL https://download.docker.com/linux/$(. /etc/os-release; echo "$ID")/gpg \
  | apt-key add - && \
  add-apt-repository \
  "deb [arch=amd64] https://download.docker.com/linux/$(. /etc/os-release; echo "$ID") \
  $(lsb_release -cs) stable"

ENV DOCKER_VERSION=18.06.1*

RUN apt-get update && \
  apt-get install -y --no-install-recommends docker-ce=${DOCKER_VERSION} && \
  sed -i 's/cgroupfs_mount$/#cgroupfs_mount\n/' /etc/init.d/docker

# Move Docker's storage location
RUN echo 'DOCKER_OPTS="${DOCKER_OPTS} --data-root=/docker-graph"' | \
  tee --append /etc/default/docker
RUN mkdir /docker-graph


RUN apt-get install dirmngr
RUN apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys 379CE192D401AB61
RUN echo "deb https://dl.bintray.com/loadimpact/deb stable main" | tee -a /etc/apt/sources.list
RUN apt-get update && apt-get install k6 && rm -rf /var/lib/apt/lists/*

# Install Helm
ENV HELM_VERSION=dev-v3

ENV HELM_FILE_NAME=helm-${HELM_VERSION}-linux-amd64.tar.gz
ENV HELM_URL=https://storage.googleapis.com/kubernetes-helm/${HELM_FILE_NAME}

RUN curl -o /tmp/$HELM_FILE_NAME ${HELM_URL} \
  && tar -zxvf /tmp/${HELM_FILE_NAME} -C /tmp \
  && mv /tmp/linux-amd64/helm /usr/local/bin/helm \
  && rm -rf /tmp/linux-amd64/helm

RUN mkdir /test-infra

COPY prow/scripts/library.sh /test-infra/prow/scripts/library.sh
COPY prow/scripts/cluster-integration/helpers/aks-library.sh /test-infra/prow/scripts/cluster-integration/helpers/aks-library.sh
COPY prow/scripts/cluster-integration/helpers/cleanup-cluster.sh /test-infra/prow/scripts/cluster-integration/helpers/cleanup-cluster.sh
COPY prow/scripts/cluster-integration/helpers/create-cluster.sh /test-infra/prow/scripts/cluster-integration/helpers/create-cluster.sh
COPY prow/scripts/cluster-integration/helpers/create-image.sh /test-infra/prow/scripts/cluster-integration/helpers/create-image.sh
COPY prow/scripts/cluster-integration/helpers/delete-image.sh /test-infra/prow/scripts/cluster-integration/helpers/delete-image.sh
COPY prow/scripts/cluster-integration/helpers/deprovision-gke-cluster.sh /test-infra/prow/scripts/cluster-integration/helpers/deprovision-gke-cluster.sh
COPY prow/scripts/cluster-integration/helpers/get-helm-certs.sh /test-infra/prow/scripts/cluster-integration/helpers/get-helm-certs.sh
COPY prow/scripts/cluster-integration/helpers/install-kyma.sh /test-infra/prow/scripts/cluster-integration/helpers/install-kyma.sh
COPY prow/scripts/cluster-integration/helpers/provision-gke-cluster.sh /test-infra/prow/scripts/cluster-integration/helpers/provision-gke-cluster.sh
COPY prow/scripts/cluster-integration/helpers/create-config-map.sh /test-infra/prow/scripts/cluster-integration/helpers/create-config-map.sh

RUN mkdir /perf-tests

COPY performance-tools/performance-cluster/runner.sh /perf-tests/runner.sh
COPY performance-tools/performance-cluster/cluster.sh /perf-tests/cluster.sh
COPY performance-tools/performance-cluster/scripts/k6-runner.sh /perf-tests/scripts/k6-runner.sh
COPY performance-tools/performance-cluster/scripts/kyma-gke-cluster.sh /perf-tests/scripts/kyma-gke-cluster.sh
COPY performance-tools/performance-cluster/scripts/kyma-aks-cluster.sh /perf-tests/scripts/kyma-aks-cluster.sh

WORKDIR /perf-tests

CMD ["./runner.sh"]
