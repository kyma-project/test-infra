FROM golang:1.19.1-alpine3.16

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . .

RUN  CGO_ENABLED=0 go build -o /gardener-rotate -ldflags="-s -w" development/gardener-rotate/main.go && \
    chmod +x /gardener-rotate
ENTRYPOINT ["/gardener-rotate"]
