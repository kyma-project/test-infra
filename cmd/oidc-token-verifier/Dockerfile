FROM golang:1.22.3-alpine as builder

# Add certificate authorities certificates.
# This let oidctokenverifier to use tls connection to the OIDC provider.
RUN apk update && apk upgrade && apk add --no-cache ca-certificates
RUN update-ca-certificates

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . .

RUN  CGO_ENABLED=0 go build -o /oidctokenverifier -ldflags="-s -w" ./cmd/oidc-token-verifier

FROM scratch

COPY --from=builder /oidctokenverifier /oidctokenverifier
# Add certificate authorities certificates.
# This let oidctokenverifier to use tls connection to the OIDC provider.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/oidctokenverifier"]