FROM golang:1.19.4-alpine3.17 AS builder

COPY . /test-infra
WORKDIR /test-infra
RUN go mod tidy && \
    go get github.com/onsi/ginkgo/v2/ginkgo/generators && \
    go get github.com/onsi/gomega/... && \
    go install github.com/onsi/ginkgo/v2/ginkgo

FROM golang:1.19.4-alpine3.17

COPY --from=builder /go/bin/ginkgo /ginkgo
RUN apk add --no-cache ca-certificates && \
    chmod +x /ginkgo
ENTRYPOINT ["/ginkgo"]
