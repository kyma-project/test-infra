FROM golang:1.20.0-alpine3.17 AS builder

# Commit details

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . .
RUN CGO_ENABLED=0 go build -o /automatedapprover -ldflags="-s -w" ./development/external-plugins/automated-approver

FROM alpine:3.17.1

COPY --from=builder /automatedapprover /automatedapprover
RUN apk add --no-cache ca-certificates git && \
    chmod  a+x /automatedapprover
ENTRYPOINT ["/automatedapprover"]
