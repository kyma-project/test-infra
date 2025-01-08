FROM golang:1.23.4-alpine3.21 as builder

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . .

RUN  CGO_ENABLED=0 go build -o /rotateserviceaccount -ldflags="-s -w" ./cmd/cloud-run/rotate-service-account

FROM alpine:3.21.1

COPY --from=builder /rotateserviceaccount /rotateserviceaccount
RUN apk add --no-cache ca-certificates && \
	chmod +x /rotateserviceaccount
ENTRYPOINT ["/rotateserviceaccount"]
