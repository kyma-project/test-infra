FROM golang:1.23.1-alpine3.20 as builder

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . .

RUN  CGO_ENABLED=0 go build -o /serviceaccountkeyscleaner -ldflags="-s -w" ./cmd/cloud-run/service-account-keys-cleaner

FROM alpine:3.20.2

COPY --from=builder /serviceaccountkeyscleaner /serviceaccountkeyscleaner
RUN apk add --no-cache ca-certificates && \
	chmod +x /serviceaccountkeyscleaner
ENTRYPOINT ["/serviceaccountkeyscleaner"]
