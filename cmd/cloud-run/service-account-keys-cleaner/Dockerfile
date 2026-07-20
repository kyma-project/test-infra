FROM europe-docker.pkg.dev/kyma-project/prod/external/library/golang:1.26.5-alpine3.23 as builder

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . .

RUN  CGO_ENABLED=0 go build -o /serviceaccountkeyscleaner -ldflags="-s -w" ./cmd/cloud-run/service-account-keys-cleaner

FROM europe-docker.pkg.dev/kyma-project/prod/external/library/alpine:3.24.1

COPY --from=builder /serviceaccountkeyscleaner /serviceaccountkeyscleaner
RUN apk add --no-cache ca-certificates && \
	chmod +x /serviceaccountkeyscleaner
ENTRYPOINT ["/serviceaccountkeyscleaner"]
