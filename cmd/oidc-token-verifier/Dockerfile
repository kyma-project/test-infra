FROM golang:1.22.2-alpine as builder

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . .

RUN  CGO_ENABLED=0 go build -o /oidctokenverifier -ldflags="-s -w" ./cmd/oidc-token-verifier

FROM scratch

COPY --from=builder /oidctokenverifier /oidctokenverifier
ENTRYPOINT ["/oidctokenverifier"]