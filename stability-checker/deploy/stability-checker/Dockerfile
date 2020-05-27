FROM ubuntu:16.04

RUN apt-get update && apt-get install -y \
    curl \
    jq \
    zip

# Install kubectl
RUN curl -Lo /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v1.17.3/bin/linux/amd64/kubectl
RUN chmod +x /usr/local/bin/kubectl

# Install KymaCLI
RUN export KYMA_RELEASE=$(curl -s https://api.github.com/repos/kyma-project/cli/releases/latest | grep tag_name | cut -d '"' -f 4) \
&& curl -Lo kyma.tar.gz "https://github.com/kyma-project/cli/releases/download/${KYMA_RELEASE}/kyma_Linux_x86_64.tar.gz" \
&& mkdir kyma-release && tar -C kyma-release -zxvf kyma.tar.gz && chmod +x kyma-release/kyma && mv kyma-release/kyma /usr/local/bin \
&& rm -rf kyma-release kyma.tar.gz

# Binaries
COPY ./stability-checker /root/stability-checker
COPY ./logs-printer /bin/logs-printer

# Scripts
COPY ./testing.sh /data/input/testing.sh
COPY ./testing-helpers.sh /data/input/testing-helpers.sh
COPY ./log.sh /data/input/log.sh

LABEL source=git@github.com:kyma-project/kyma.git

ENTRYPOINT ["/root/stability-checker"]