# Basic golang buildpack

FROM eu.gcr.io/kyma-project/test-infra/bootstrap:v20200124-900eb728

#################################################################
#################### Add commit details #########################
#################################################################
ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit
#################################################################

# Versions

ENV GO_VERSION 1.14.2
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

#################################################################
########################## Prow Tools ###########################
#################################################################

COPY --from=eu.gcr.io/kyma-project/test-infra/prow-tools:v20200827-b59d2d67 /prow-tools /prow-tools
# for better access to prow-tools
ENV PATH=$PATH:/prow-tools
