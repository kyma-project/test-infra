FROM golang:1.19.2-alpine3.16 as builder

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . .

RUN  CGO_ENABLED=0 go build -o /kms-rotate -ldflags="-s -w" ./development/gcp-cloud-run/kms-rotate

FROM alpine:3.16.2

COPY --from=builder /kms-rotate /kms-rotate
RUN apk add --no-cache ca-certificates && \
	chmod +x /kms-rotate
ENTRYPOINT ["/kms-rotate"]
