FROM golang:1.14.1-alpine3.11 AS builder

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . development/tools

RUN apk --update --no-cache add bash upx dep git && \
    development/tools/build-cleaners.sh

FROM alpine:latest

RUN apk --update --no-cache add ca-certificates bash
COPY --from=builder /go/src/github.com/kyma-project/test-infra/development/tools/bin /tools
WORKDIR /tools