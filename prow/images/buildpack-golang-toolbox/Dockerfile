# Basic golang buildpack

FROM eu.gcr.io/kyma-project/prow/test-infra/buildpack-golang:go1.12

# Install additional tools

RUN go get golang.org/x/tools/cmd/goimports
RUN go get golang.org/x/lint/golint
RUN go get github.com/vektra/mockery
RUN go get github.com/kyma-project/kyma/tools/failery
RUN go get github.com/kisielk/errcheck