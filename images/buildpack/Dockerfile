FROM alpine:3.18.2
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
    dumb-init \
    libgit2-dev \
    pkgconf

RUN adduser -D prow