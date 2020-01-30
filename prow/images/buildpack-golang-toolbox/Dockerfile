# Basic golang buildpack

FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.13

# Install additional tools

RUN apt-get update && apt-get install -y --no-install-recommends \
    rsync=3.1.2-1+deb9u2 \
    procps=2:3.3.12-3+deb9u1 \
    && apt-get clean

RUN go get golang.org/x/tools/cmd/goimports  \
    golang.org/x/lint/golint \
    github.com/vektra/mockery \
    github.com/kisielk/errcheck && \
    GO111MODULE="on" go get sigs.k8s.io/kind@v0.5.1

COPY ./license-puller.sh /license-puller.sh
ENV LICENSE_PULLER_PATH=/license-puller.sh
