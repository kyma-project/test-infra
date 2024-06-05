FROM golang:1.22.4-alpine as builder

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . .

RUN  CGO_ENABLED=0 go build -o /serviceaccountkeyscleaner -ldflags="-s -w" ./cmd/cloud-run/service-account-keys-cleaner

FROM alpine:3.20.0

COPY --from=builder /serviceaccountkeyscleaner /serviceaccountkeyscleaner
RUN apk add --no-cache ca-certificates && \
	chmod +x /serviceaccountkeyscleaner
ENTRYPOINT ["/serviceaccountkeyscleaner"]
