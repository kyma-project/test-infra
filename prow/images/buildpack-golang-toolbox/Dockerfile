# Basic golang buildpack

FROM eu.gcr.io/kyma-project/prow/test-infra/bootstrap:v20181204-a6e79be

# Versions

ENV GO_VERSION 1.12.7
ENV DEP_RELEASE_TAG v0.5.4

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

# Install additional tools

RUN go get golang.org/x/tools/cmd/goimports
RUN go get golang.org/x/lint/golint
RUN go get github.com/vektra/mockery
RUN go get github.com/kyma-project/kyma/tools/failery
RUN go get github.com/kisielk/errcheck