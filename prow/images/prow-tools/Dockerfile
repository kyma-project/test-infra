FROM golang:1.14.1-alpine3.11 AS builder

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . .

RUN apk --update --no-cache add bash upx dep git && \
    development/tools/build-cleaners.sh && \
    cd development/jobguard/ && CGO_ENABLED=0 go build -o /prow-tools/jobguard -ldflags="-s -w" cmd/jobguard/main.go

FROM alpine:latest

RUN apk --update --no-cache add ca-certificates bash
COPY --from=builder /go/src/github.com/kyma-project/test-infra/development/tools/bin /prow-tools
COPY --from=builder /prow-tools/jobguard /prow-tools
WORKDIR /prow-tools