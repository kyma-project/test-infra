FROM golang:1.20.2-alpine3.17 as builder

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . .

RUN  CGO_ENABLED=0 go build -o /scanlogsforsecrets -ldflags="-s -w" ./development/gcp/cloud-run/scan-logs-for-secrets

FROM alpine:3.17.2

COPY --from=builder /scanlogsforsecrets /scanlogsforsecrets
COPY --from=builder /go/src/github.com/kyma-project/test-infra/development/gcp/cloud-run/scan-logs-for-secrets/gitleaks.toml /app/

RUN apk add --no-cache ca-certificates && \
	chmod +x /scanlogsforsecrets
ENTRYPOINT ["/scanlogsforsecrets"]
