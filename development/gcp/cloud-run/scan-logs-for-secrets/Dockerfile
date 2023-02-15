FROM golang:1.20.0-alpine3.17 as builder

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . .
RUN ls -alh
#RUN  CGO_ENABLED=0 go build -o /scanlogsforsecrets -ldflags="-s -w" ./development/gcp/cloud-run/scan-logs-for-secrets

FROM alpine:3.17.1

#COPY --from=builder /scanlogsforsecrets /scanlogsforsecrets
COPY --from=builder /go/src/github.com/kyma-project/test-infra /app/
RUN ls -alh
RUN ls /app/ -alh
#RUN apk add --no-cache ca-certificates && \
#	chmod +x /scanlogsforsecrets
#ENTRYPOINT ["/scanlogsforsecrets"]
