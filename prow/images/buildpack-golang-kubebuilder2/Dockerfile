# Golang buildpack with kubebuilder

FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20200117-d3885041

# Buildpack variables

ARG BUILDPACK_VERSION=latest
ENV BUILDPACK_VERSION ${BUILDPACK_VERSION}

# Versions

ENV KUBEBUILDER_VERSION 2.0.0
ENV KUSTOMIZE_VERSION 3.1.0
ENV ARCH amd64

# Install kubebuilder

RUN curl -L -O "https://github.com/kubernetes-sigs/kubebuilder/releases/download/v${KUBEBUILDER_VERSION}/kubebuilder_${KUBEBUILDER_VERSION}_linux_${ARCH}.tar.gz" && \
    tar -zxvf kubebuilder_${KUBEBUILDER_VERSION}_linux_${ARCH}.tar.gz && \
    rm kubebuilder_${KUBEBUILDER_VERSION}_linux_${ARCH}.tar.gz && \
    mv kubebuilder_${KUBEBUILDER_VERSION}_linux_${ARCH} kubebuilder && \
    mv kubebuilder /usr/local/

RUN curl -L -o kustomize "https://github.com/kubernetes-sigs/kustomize/releases/download/v${KUSTOMIZE_VERSION}/kustomize_${KUSTOMIZE_VERSION}_linux_${ARCH}" && \
    chmod u+x kustomize && \
    mv kustomize /usr/local/bin/kustomize

# Paths

ENV PATH /usr/local/kubebuilder/bin:$PATH
# Path used by decorator
ENV PATH /home/prow/go/bin:$PATH
