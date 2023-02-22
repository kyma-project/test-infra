FROM golang:1.20.1-alpine3.17 as builder

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . .

RUN  CGO_ENABLED=0 go build -o /rotateserviceaccount -ldflags="-s -w" ./development/secrets-rotator/cloud-run/rotate-service-account

FROM alpine:3.17.2

COPY --from=builder /rotateserviceaccount /rotateserviceaccount
RUN apk add --no-cache ca-certificates && \
	chmod +x /rotateserviceaccount
ENTRYPOINT ["/rotateserviceaccount"]
