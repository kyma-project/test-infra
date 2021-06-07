# Basic golang buildpack

FROM eu.gcr.io/kyma-project/test-infra/bootstrap:v20200124-900eb728

# Commit details

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

# Versions

ENV ARCH amd64
ENV GO_VERSION=1.16.4
ENV DEP_RELEASE_TAG=v0.5.4

ENV KUBEBUILDER_VERSION 2.3.1
ENV KUSTOMIZE_VERSION 3.8.4

# Install Go

RUN wget -q https://dl.google.com/go/go${GO_VERSION}.linux-${ARCH}.tar.gz && \
    tar xzf go${GO_VERSION}.linux-${ARCH}.tar.gz && \
    rm go${GO_VERSION}.linux-${ARCH}.tar.gz && \
    mv go /usr/local

ENV GOPATH /workspace/go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p ${GOPATH}/bin && \
    mkdir -p ${GOPATH}/src

#Install kubebuilder
RUN curl -L -O "https://github.com/kubernetes-sigs/kubebuilder/releases/download/v${KUBEBUILDER_VERSION}/kubebuilder_${KUBEBUILDER_VERSION}_linux_${ARCH}.tar.gz" && \
    tar -zxvf kubebuilder_${KUBEBUILDER_VERSION}_linux_${ARCH}.tar.gz && \
    rm kubebuilder_${KUBEBUILDER_VERSION}_linux_${ARCH}.tar.gz && \
    mv kubebuilder_${KUBEBUILDER_VERSION}_linux_${ARCH} kubebuilder && \
    mv kubebuilder /usr/local/ && \
    curl -L -O "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize/v${KUSTOMIZE_VERSION}/kustomize_v${KUSTOMIZE_VERSION}_linux_${ARCH}.tar.gz" && \
    tar -zxvf kustomize_v${KUSTOMIZE_VERSION}_linux_${ARCH}.tar.gz && \
    rm kustomize_v${KUSTOMIZE_VERSION}_linux_${ARCH}.tar.gz && \
    mv kustomize /usr/local/bin/kustomize

# Paths

ENV PATH /usr/local/kubebuilder/bin:$PATH
# Path used by decorator
ENV PATH /home/prow/go/bin:$PATH

# Install Dep

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

# Prow Tools

COPY --from=eu.gcr.io/kyma-project/test-infra/prow-tools:v20210401-294e46e5 /prow-tools /prow-tools
# for better access to prow-tools
ENV PATH=$PATH:/prow-tools
