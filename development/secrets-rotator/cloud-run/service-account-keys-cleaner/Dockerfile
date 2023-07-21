FROM golang:1.20.0-alpine3.17 as builder

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . .

RUN  CGO_ENABLED=0 go build -o /serviceaccountkeyscleaner -ldflags="-s -w" ./development/secrets-rotator/cloud-run/service-account-keys-cleaner

FROM alpine:3.17.2

COPY --from=builder /serviceaccountkeyscleaner /serviceaccountkeyscleaner
RUN apk add --no-cache ca-certificates && \
	chmod +x /serviceaccountkeyscleaner
ENTRYPOINT ["/serviceaccountkeyscleaner"]
