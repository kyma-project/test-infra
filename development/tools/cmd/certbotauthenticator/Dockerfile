FROM golang:1.14.1-alpine3.11 AS builder

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . .

RUN apk --update --no-cache add bash upx dep git && \
    cd development/gcp/cmd/certbotauthenticator && CGO_ENABLED=0 go build -o ./certbotauthenticator -ldflags="-s -w" ./main.go