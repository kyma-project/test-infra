#FROM golang:1.19.3-alpine3.16 as builder

#WORKDIR /go/src/github.com/kyma-project/test-infra
#COPY . .

#RUN  CGO_ENABLED=0 go build -o /scanlogsforsecrets -ldflags="-s -w" ./development/gcp/cloud-run/scan-logs-for-secrets

FROM alpine:3.16.2

#COPY --from=builder /scanlogsforsecrets /scanlogsforsecrets
COPY scanlogsforsecrets /scanlogsforsecrets
RUN apk add --no-cache ca-certificates && \
	chmod +x /scanlogsforsecrets
ENTRYPOINT ["/scanlogsforsecrets"]
