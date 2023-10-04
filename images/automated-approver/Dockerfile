FROM golang:1.21-alpine3.18 AS builder

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . .
RUN CGO_ENABLED=0 go build -o /automatedapprover -ldflags="-s -w" ./cmd/automated-approver

FROM alpine:3.18

COPY --from=builder /automatedapprover /automatedapprover
RUN apk upgrade --no-cache && apk add --no-cache ca-certificates git && update-ca-certificates \
    chmod  a+x /automatedapprover
ENTRYPOINT ["/automatedapprover"]
