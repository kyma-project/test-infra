FROM golang:1.14.1-buster AS builder

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . development/tools

RUN apt-get update && apt-get install -y git upx && \
    curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh && \
    development/tools/build-cleaners.sh

FROM debian:10

RUN apt-get update && apt-get upgrade -y
COPY --from=builder /go/src/github.com/kyma-project/test-infra/development/tools/bin /prow-tools
WORKDIR /prow-tools