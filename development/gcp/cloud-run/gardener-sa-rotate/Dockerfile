FROM golang:1.19.4-alpine3.17 as builder

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . .

RUN  CGO_ENABLED=0 go build -o /gardener-sa-rotate -ldflags="-s -w" ./development/gcp/cloud-run/gardener-sa-rotate

FROM alpine:3.16.2

COPY --from=builder /gardener-sa-rotate /gardener-sa-rotate
RUN apk add --no-cache ca-certificates && \
	chmod +x /gardener-sa-rotate
ENTRYPOINT ["/gardener-sa-rotate"]
