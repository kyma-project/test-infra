# Basic golang buildpack

FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.12

# Install additional tools

RUN go get golang.org/x/tools/cmd/goimports
RUN go get golang.org/x/lint/golint
RUN go get github.com/vektra/mockery
RUN go get github.com/kisielk/errcheck

COPY ./licence-puller.sh /licence-puller.sh
ENV LICENSE_PULLER_PATH=/licence-puller.sh