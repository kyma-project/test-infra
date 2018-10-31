# Basic golang buildpack

FROM eu.gcr.io/kyma-project/prow/bootstrap:0.0.1

# Versions

ENV GO_VERSION 1.10.4
ENV DEP_RELEASE_TAG v0.4.1

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