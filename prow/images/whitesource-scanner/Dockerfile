FROM eu.gcr.io/kyma-project/test-infra/bootstrap:v20221011-310d4b98

# Commit details
ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

ENV UA_VERSION v22.1.1.1
ENV GO_VERSION 1.19.1
ENV DEP_RELEASE_TAG v0.5.4
ENV JAVA_VERSION 8u212-b04
ENV JAVA_BASE_URL https://github.com/AdoptOpenJDK/openjdk8-binaries/releases/download/jdk8u212-b04/OpenJDK8U-jdk_
ENV JAVA_URL_VERSION 8u212b04

ENV JAVA_HOME /usr/local/openjdk-8
ENV PATH $JAVA_HOME/bin:$PATH

# Default to UTF-8 file.encoding
ENV LANG C.UTF-8

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates p11-kit maven \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fLSs -o openjdk.tgz "${JAVA_BASE_URL}x64_linux_hotspot_${JAVA_URL_VERSION}.tar.gz"

RUN mkdir -p "$JAVA_HOME"; \
	tar --extract \
		--file openjdk.tgz \
		--directory "$JAVA_HOME" \
		--strip-components 1 \
		--no-same-owner \
	; \
	rm openjdk.tgz* \
	; \
    java -version


# Install Go
ENV GOPATH /workspace/go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN curl -o go.tar.gz -LfSs https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz && \
    tar xzf go.tar.gz && \
    rm go.tar.gz && \
    mv go /usr/local && \
    mkdir -p ${GOPATH}/bin && \
    mkdir -p ${GOPATH}/src


# Install Dep
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

# install unified agent
RUN curl -LJO https://github.com/whitesource/unified-agent-distribution/releases/download/"${UA_VERSION}"/wss-unified-agent.jar && \
    mkdir -p /wss && \
    mv wss-unified-agent.jar /wss/wss-unified-agent.jar

# install NodeJS
RUN curl -sL https://deb.nodesource.com/setup_12.x | bash -
RUN apt-get update && apt-get install -y --no-install-recommends \
    nodejs \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Prow Tools
# hadolint doesn't support --from external image
# hadolint ignore=DL3022
COPY --from=eu.gcr.io/kyma-project/test-infra/prow-tools:v20210401-294e46e5 /prow-tools /prow-tools
# for better access to prow-tools
ENV PATH=$PATH:/prow-tools
