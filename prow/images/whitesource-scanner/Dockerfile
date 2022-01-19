FROM eu.gcr.io/kyma-project/test-infra/bootstrap:v20211020-638b6f95

# Commit details
ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

ENV UA_VERSION v20.8.2
ENV GO_VERSION 1.17.5
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
RUN curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add -
RUN echo "deb https://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list

RUN curl -sL https://deb.nodesource.com/setup_12.x | bash -

RUN apt-get update && apt-get install -y --no-install-recommends \
    libfontconfig1 \
    procps \
    nodejs \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# RUN npm install -g eslint-config-react-app@^3.0.4 \
#     babel-eslint@^9.0.0 \
#     eslint@^5.6.1 \
#     eslint-plugin-flowtype@^2.50.3 \
#     eslint-plugin-import@^2.14.0 \
#     eslint-plugin-jsx-a11y@^6.1.2 \
#     eslint-plugin-react@^7.11.1 \
#     tslint@^5.11.0 \
#     tslint-angular@^1.1.2 \
#     tslint-config-prettier@^1.15.0 \
#     typescript@^3.1.3 \
#     prettier@^1.14.3

# Prow Tools
# hadolint doesn't support --from external image
# hadolint ignore=DL3022
COPY --from=eu.gcr.io/kyma-project/test-infra/prow-tools:v20210401-294e46e5 /prow-tools /prow-tools
# for better access to prow-tools
ENV PATH=$PATH:/prow-tools
