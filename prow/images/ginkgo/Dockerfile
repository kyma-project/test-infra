FROM golang:1.19.4-alpine3.17 AS builder

RUN go install github.com/onsi/ginkgo/v2/ginkgo && \
    go get github.com/onsi/gomega/...
RUN apk add --no-cache ca-certificates && \
	chmod +x /gardener-rotate \

ENTRYPOINT ["ginkgo"]
