FROM alpine:3.20.1

ARG DOCKER_CREDENTIAL_GCR_VERSION=2.1.22
RUN set -eux; \
    apk upgrade --no-cache && \
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
    binutils-gold \
    dumb-init \
    libgit2-dev \
    pkgconf \
    xz>5.6.2-r0 \
    && \
    arch="$(apk --print-arch)"; \
    case "$arch" in \
        'x86_64') \
            export GOAMD64='v1' GOARCH='amd64' GOOS='linux'; \
            ;; \
        'armhf') \
            export GOARCH='arm' GOARM='6' GOOS='linux'; \
            ;; \
        'armv7') \
            export GOARCH='arm' GOARM='7' GOOS='linux'; \
            ;; \
        'aarch64') \
            export GOARCH='arm64' GOOS='linux'; \
            ;; \
        'x86') \
            export GO386='softfloat' GOARCH='386' GOOS='linux'; \
            ;; \
        'ppc64le') \
            export GOARCH='ppc64le' GOOS='linux'; \
            ;; \
        'riscv64') \
            export GOARCH='riscv64' GOOS='linux'; \
            ;; \
        's390x') \
            export GOARCH='s390x' GOOS='linux'; \
            ;; \
        *) echo >&2 "error: unsupported architecture '$arch' (likely packaging update needed)"; exit 1 ;; \
    esac; \
    curl -fsSLo docker-credential-gcr.tar.gz "https://github.com/GoogleCloudPlatform/docker-credential-gcr/releases/download/v${DOCKER_CREDENTIAL_GCR_VERSION}/docker-credential-gcr_${GOOS}_${GOARCH}-${DOCKER_CREDENTIAL_GCR_VERSION}.tar.gz" && \
    tar xzf docker-credential-gcr.tar.gz \
    && chmod +x docker-credential-gcr && mv docker-credential-gcr /usr/bin/ \
    && docker-credential-gcr configure-docker --registries=eu.gcr.io,europe-docker.pkg.dev; \
    adduser -D prow
