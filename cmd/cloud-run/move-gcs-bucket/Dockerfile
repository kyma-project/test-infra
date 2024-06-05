FROM golang:1.22.4-alpine as builder

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . .

RUN  CGO_ENABLED=0 go build -o /movegcsbucket -ldflags="-s -w" ./cmd/cloud-run/move-gcs-bucket

FROM alpine:3.20.0

COPY --from=builder /movegcsbucket /movegcsbucket
RUN apk add --no-cache ca-certificates && \
	chmod +x /movegcsbucket
ENTRYPOINT ["/movegcsbucket"]
