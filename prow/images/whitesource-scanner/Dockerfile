FROM eu.gcr.io/kyma-project/test-infra/buildpack-java-node:v20200226-b9357193

# Commit details
ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

ENV UA_VERSION v20.8.2
ENV GO_VERSION 1.14.2
ENV DEP_RELEASE_TAG v0.5.4

# Install Go
ENV GOPATH /workspace/go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN wget -q https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz && \
    tar xzf go${GO_VERSION}.linux-amd64.tar.gz && \
    rm go${GO_VERSION}.linux-amd64.tar.gz && \
    mv go /usr/local && \

    mkdir -p ${GOPATH}/bin && \
    mkdir -p ${GOPATH}/src



# Install Dep
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

# install unified agent
RUN curl -LJO https://github.com/whitesource/unified-agent-distribution/releases/download/"${UA_VERSION}"/wss-unified-agent.jar && \
    mkdir -p /wss && \
    mv wss-unified-agent.jar /wss/wss-unified-agent.jar

# Prow Tools
COPY --from=eu.gcr.io/kyma-project/test-infra/prow-tools:v20201008-a2fd337c /prow-tools /prow-tools
# for better access to prow-tools
ENV PATH=$PATH:/prow-tools
