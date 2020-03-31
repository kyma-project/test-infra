FROM golang:alpine AS builder

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . development/tools

RUN apk --update --no-cache add bash upx dep git && \
    development/tools/build-cleaners.sh

FROM alpine:latest

RUN apk --update --no-cache add ca-certificates bash
COPY --from=builder /go/src/github.com/kyma-project/test-infra/development/tools/bin /tools
ENV PATH $PATH:/tools