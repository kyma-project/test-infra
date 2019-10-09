# Basic golang buildpack

FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.12

# Install additional tools

RUN apt-get update && apt-get install -y --no-install-recommends \
    rsync \
    procps \
    && apt-get clean

RUN go get golang.org/x/tools/cmd/goimports  \
    golang.org/x/lint/golint \
    github.com/vektra/mockery \
    github.com/kisielk/errcheck && \
    GO111MODULE="on" go get sigs.k8s.io/kind@v0.5.1

# Increase watches and instances
RUN echo fs.inotify.max_user_watches=524288 | tee -a /etc/sysctl.conf && sysctl -p \
    && echo fs.inotify.max_user_instances=512 | tee -a /etc/sysctl.conf && sysctl -p

COPY ./license-puller.sh /license-puller.sh
ENV LICENSE_PULLER_PATH=/license-puller.sh
