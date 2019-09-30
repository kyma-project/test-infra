FROM eu.gcr.io/kyma-project/prow/test-infra/buildpack-golang:v20181204-a6e79be

# Install packages required by az
RUN apt-get update && apt-get install -y --no-install-recommends \
    apt-transport-https \
    lsb-release \
    software-properties-common \
    dirmngr \
    gnupg2

ENV HELM_VERSION="v2.10.0"
ENV AZURE_CLI_VERSION="2.0.62-1~stretch"

RUN wget -q https://storage.googleapis.com/kubernetes-helm/helm-${HELM_VERSION}-linux-amd64.tar.gz -O - | tar -xzO linux-amd64/helm > /usr/local/bin/helm \
    && chmod +x /usr/local/bin/helm

RUN helm init --client-only

# Install az
# If you receive error related to recv-keys please make sure you have latest signing key
# https://docs.microsoft.com/en-us/cli/azure/install-azure-cli-apt?view=azure-cli-latest#signingKey
RUN echo "deb [arch=amd64] https://packages.microsoft.com/repos/azure-cli/ "$(lsb_release -cs)" main" | \
    tee /etc/apt/sources.list.d/azure-cli.list && \
    apt-key --keyring /etc/apt/trusted.gpg.d/Microsoft.gpg adv \
        --keyserver packages.microsoft.com \
        --recv-keys BC528686B50D79E339D3721CEB3E94ADBE1229CF && \
    apt-get update && \
    apt-get install -y azure-cli=${AZURE_CLI_VERSION}

# Install kubectl
ENV KUBECTL_VERSION="v1.14.6"
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl --fail &&\
	chmod +x kubectl &&\
	mv kubectl /usr/local/bin/kubectl
