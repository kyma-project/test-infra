FROM golang:1.20.0-alpine3.17 AS builder

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . .

RUN  CGO_ENABLED=0 go build -o /gardener-rotate -ldflags="-s -w" ./development/gardener-rotate


FROM alpine:3.17.1

COPY --from=builder /gardener-rotate /gardener-rotate
RUN apk add --no-cache ca-certificates && \
	chmod +x /gardener-rotate
ENTRYPOINT ["/gardener-rotate"]
