FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang:v20200124-deef6b8c

# Commit details

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

# Install packages required by az
RUN apt-get update && apt-get install -y --no-install-recommends \
    apt-transport-https \
    lsb-release \
    software-properties-common \
    dirmngr \
    gnupg2

ENV HELM_VERSION="v3.2.1"
ENV AZURE_CLI_VERSION="2.0.62-1~stretch"

RUN wget https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz -O - | tar -xzO linux-amd64/helm > /usr/local/bin/helm &&\
	  chmod +x /usr/local/bin/helm &&\
      rm -rf helm-${HELM_VERSION}-linux-amd64.tar.gz linux-amd64

# Install az
# If you receive error related to recv-keys please make sure you have latest signing key
# https://docs.microsoft.com/en-us/cli/azure/install-azure-cli-apt?view=azure-cli-latest#signingKey
RUN echo "deb [arch=amd64] https://packages.microsoft.com/repos/azure-cli/ "$(lsb_release -cs)" main" | \
    tee /etc/apt/sources.list.d/azure-cli.list && \
    curl -sL "https://packages.microsoft.com/keys/microsoft.asc" | apt-key add - && \
    apt-get update && \
    apt-get install -y azure-cli=${AZURE_CLI_VERSION}

# Prow Tools

COPY --from=eu.gcr.io/kyma-project/test-infra/prow-tools:v20200827-b59d2d67 /prow-tools /prow-tools
# for better access to prow-tools
ENV PATH=$PATH:/prow-tools