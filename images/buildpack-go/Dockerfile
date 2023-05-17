FROM golang:1.20.4-alpine as base
RUN set -eux; \
    apk add --no-cache \
    bash \
    curl \
    git \
    build-base \
    jq \
    wget \
    ca-certificates \
    gettext \
    yq \
    binutils \
    dumb-init

RUN adduser -D prow

FROM base AS deps

ARG KUSTOMIZE_VERSION=v5.0.3
ARG KUBEBUILDER_VERSION=v3.10.0
ARG KO_VERSION=cbb93deee1953c1e9d5f974abd62a96676aa68f9
ARG DOCKER_CREDENTIAL_GCR_VERSION=2.1.8

RUN set -eux; \
    curl -L -o kubebuilder "https://github.com/kubernetes-sigs/kubebuilder/releases/download/${KUBEBUILDER_VERSION}/kubebuilder_$(go env GOOS)_$(go env GOARCH)" && \
    chmod +x kubebuilder && mv kubebuilder /go/bin/ && \
    CGO_ENABLED=0 go install "github.com/google/ko@${KO_VERSION}" && \
    CGO_ENABLED=0 go install "sigs.k8s.io/kustomize/kustomize/v5@${KUSTOMIZE_VERSION}"

RUN curl -fsSLo docker-credential-gcr.tar.gz "https://github.com/GoogleCloudPlatform/docker-credential-gcr/releases/download/v${DOCKER_CREDENTIAL_GCR_VERSION}/docker-credential-gcr_$(go env GOOS)_$(go env GOARCH)-${DOCKER_CREDENTIAL_GCR_VERSION}.tar.gz" && \
    tar xzf docker-credential-gcr.tar.gz \
    && chmod +x docker-credential-gcr && mv docker-credential-gcr /usr/bin/

FROM base AS tools

RUN CGO_ENABLED=0 go install github.com/kyma-project/test-infra/development/jobguard/cmd/jobguard@main

FROM base

COPY --from=deps /go/bin/ko /usr/local/bin/ko
COPY --from=deps /go/bin/kubebuilder /usr/local/bin/kubebuilder
COPY --from=deps /go/bin/kustomize /usr/local/bin/kustomize
COPY --from=deps /usr/bin/docker-credential-gcr /usr/bin/docker-credential-gcr
COPY --from=tools /go/bin/jobguard /usr/local/bin/jobguard
USER prow:prow

RUN docker-credential-gcr configure-docker --registries=eu.gcr.io,europe-docker.pkg.dev

ENTRYPOINT ["/usr/bin/dumb-init", "--"]