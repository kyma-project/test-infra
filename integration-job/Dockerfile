FROM debian:stretch

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        netbase \
        openssh-client \
        wget \
        python && \
    rm -rf /var/lib/apt/lists/*

## GCLOUD
ENV CLOUD_SDK_VERSION=219.0.1
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
    gcloud components install alpha beta && \
    gcloud info 

COPY ["provision_vm_and_start_kyma.sh", "vm/install_deps_and_run_kyma.sh", "/"]
ENTRYPOINT [ "/provision_vm_and_start_kyma.sh" ]