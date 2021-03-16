# Golang buildpack with kubebuilder

FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20210125-6234473e

# Commit details

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

# Buildpack variables

ARG BUILDPACK_VERSION=latest
ENV BUILDPACK_VERSION ${BUILDPACK_VERSION}

# Versions

ENV KUBEBUILDER_VERSION 2.3.1
ENV KUSTOMIZE_VERSION 3.8.4
ENV ARCH amd64

# Install kubebuilder

RUN curl -L -O "https://github.com/kubernetes-sigs/kubebuilder/releases/download/v${KUBEBUILDER_VERSION}/kubebuilder_${KUBEBUILDER_VERSION}_linux_${ARCH}.tar.gz" && \
    tar -zxvf kubebuilder_${KUBEBUILDER_VERSION}_linux_${ARCH}.tar.gz && \
    rm kubebuilder_${KUBEBUILDER_VERSION}_linux_${ARCH}.tar.gz && \
    mv kubebuilder_${KUBEBUILDER_VERSION}_linux_${ARCH} kubebuilder && \
    mv kubebuilder /usr/local/

RUN curl -L -O "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize/v${KUSTOMIZE_VERSION}/kustomize_v${KUSTOMIZE_VERSION}_linux_${ARCH}.tar.gz" && \
    tar -zxvf kustomize_v${KUSTOMIZE_VERSION}_linux_${ARCH}.tar.gz && \
    rm kustomize_v${KUSTOMIZE_VERSION}_linux_${ARCH}.tar.gz && \
    mv kustomize /usr/local/bin/kustomize

RUN apt install -y python-pip && \
	yes | pip install cgroup-utils

# Paths

ENV PATH /usr/local/kubebuilder/bin:$PATH
# Path used by decorator
ENV PATH /home/prow/go/bin:$PATH

# Prow Tools

COPY --from=eu.gcr.io/kyma-project/test-infra/prow-tools:v20201113-72ed0661 /prow-tools /prow-tools
# for better access to prow-tools
ENV PATH=$PATH:/prow-tools
