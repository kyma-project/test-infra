# Basic golang buildpack

FROM eu.gcr.io/kyma-project/test-infra/bootstrap:v20220427-9543160d

# Commit details

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

# Versions

ENV ARCH amd64
ENV GO_VERSION 1.18
ENV DEP_RELEASE_TAG v0.5.4
ENV KUBEBUILDER_VERSION 2.3.2
ENV KUSTOMIZE_VERSION 3.8.4
ENV GOLANGCI_LINT_VERSION v1.45.2

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# apt and install additional packages
RUN apt-get update && apt-get install -y --no-install-recommends \
        rsync \
        procps \
        pkg-config \
        libgit2-dev \
        && apt-get clean \
        && rm -rf /var/lib/apt/lists/*

# Install Go and Dep
ENV GOPATH /workspace/go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN curl -fLSs -o go.tar.gz "https://dl.google.com/go/go${GO_VERSION}.linux-${ARCH}.tar.gz" && \
    tar xzf go.tar.gz && \
    rm go.tar.gz && \
    mv go /usr/local && \
    mkdir -p "${GOPATH}/bin" && \
    mkdir -p "${GOPATH}/src" && \
    curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

# Install kubebuilder
ENV PATH /usr/local/kubebuilder/bin:$PATH
ENV PATH /home/prow/go/bin:$PATH

RUN curl -fLSs -o kubebuilder.tar.gz "https://github.com/kubernetes-sigs/kubebuilder/releases/download/v${KUBEBUILDER_VERSION}/kubebuilder_${KUBEBUILDER_VERSION}_linux_${ARCH}.tar.gz" && \
    tar -zxvf kubebuilder.tar.gz && \
    rm kubebuilder.tar.gz && \
    mv "kubebuilder_${KUBEBUILDER_VERSION}_linux_${ARCH}" kubebuilder && \
    mv kubebuilder /usr/local/ && \
    curl -fLSs -o kustomize.tar.gz "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize/v${KUSTOMIZE_VERSION}/kustomize_v${KUSTOMIZE_VERSION}_linux_${ARCH}.tar.gz" && \
    tar -zxvf kustomize.tar.gz && \
    rm kustomize.tar.gz && \
    mv kustomize /usr/local/bin/kustomize

# Install golang toolbox
RUN go install golang.org/x/tools/cmd/goimports@latest && \
    go install golang.org/x/lint/golint@latest && \
    go install github.com/ericchiang/pup@latest && \
    go install github.com/kisielk/errcheck@latest && \
    go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.8.2 && \
    go install golang.org/x/perf/cmd/benchstat@latest && \
    go install github.com/vektra/mockery/v2@latest && \
    go install github.com/maxbrunsfeld/counterfeiter/v6@latest && \
    curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | bash -s -- -b ${GOPATH}/bin ${GOLANGCI_LINT_VERSION}

COPY ./license-puller.sh /license-puller.sh
ENV LICENSE_PULLER_PATH=/license-puller.sh

# Prow Tools
# hadolint ignore=DL3022
COPY --from=eu.gcr.io/kyma-project/test-infra/prow-tools:v20210804-d55639f9 /prow-tools /prow-tools
# for better access to prow-tools
ENV PATH=$PATH:/prow-tools
