FROM golang:1.19.2-alpine3.16 as builder

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . .

RUN  CGO_ENABLED=0 go build -o /update-secrets-in-store -ldflags="-s -w" ./development/jenkins/update-secrets-in-store

FROM alpine:3.16.2

COPY --from=builder /update-secrets-in-store /update-secrets-in-store
RUN apk add --no-cache ca-certificates && \
	chmod +x /update-secrets-in-store
ENTRYPOINT ["/update-secrets-in-store"]
