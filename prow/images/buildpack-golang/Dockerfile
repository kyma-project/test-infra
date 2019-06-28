# Basic golang buildpack

FROM eu.gcr.io/kyma-project/test-infra/pr/bootstrap:PR-1190

# Versions

ENV GO_VERSION 1.11.4 \
    DEP_RELEASE_TAG v0.5.0 \
    BUILDPACK_NAME=buildpack-golang

# Install Go

RUN wget -q https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz && \
    tar xzf go${GO_VERSION}.linux-amd64.tar.gz && \
    rm go${GO_VERSION}.linux-amd64.tar.gz && \
    mv go /usr/local

ENV GOPATH /workspace/go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p ${GOPATH}/bin
RUN mkdir -p ${GOPATH}/src

# Install Dep

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh