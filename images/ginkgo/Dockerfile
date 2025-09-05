ARG GOLANG_VERSION

FROM golang:${GOLANG_VERSION}-alpine3.22 AS builder

COPY . /test-infra
WORKDIR /test-infra
RUN go get github.com/onsi/ginkgo/v2/ginkgo/generators && \
    go get github.com/onsi/ginkgo/v2/ginkgo@v2.23.4 && \
    go install github.com/onsi/ginkgo/v2/ginkgo && \
    go get github.com/onsi/gomega/...

FROM golang:${GOLANG_VERSION}-alpine3.22

COPY --from=builder /go/bin/ginkgo /ginkgo
RUN apk update && \
    apk add --no-cache ca-certificates git gcc musl-dev && \
    chmod +x /ginkgo
