# Golang buildpack with kubebuilder

FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang:v20190116-69aa0a1

# Buildpack variables

ARG BUILDPACK_VERSION=latest
ENV BUILDPACK_VERSION ${BUILDPACK_VERSION}

# Versions

ENV KUBEBUILDER_VERSION 1.0.7
ENV ARCH amd64

# Install kubebuilder

RUN curl -L -O "https://github.com/kubernetes-sigs/kubebuilder/releases/download/v${KUBEBUILDER_VERSION}/kubebuilder_${KUBEBUILDER_VERSION}_linux_${ARCH}.tar.gz" && \
    tar -zxvf kubebuilder_${KUBEBUILDER_VERSION}_linux_${ARCH}.tar.gz && \
    rm kubebuilder_${KUBEBUILDER_VERSION}_linux_${ARCH}.tar.gz && \
    mv kubebuilder_${KUBEBUILDER_VERSION}_linux_${ARCH} kubebuilder && \
    mv kubebuilder /usr/local/

# Paths

ENV PATH /usr/local/kubebuilder/bin:$PATH
# Path used by decorator
ENV PATH /home/prow/go/bin:$PATH
