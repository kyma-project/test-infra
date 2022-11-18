FROM golang:1.19.2-alpine3.16 as builder

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . .

RUN  CGO_ENABLED=0 go build -o /markdown-index -ldflags="-s -w" ./development/markdown-index

FROM alpine:3.16.2

COPY --from=builder /markdown-index /markdown-index
RUN apk add --no-cache ca-certificates git && \
	chmod +x /markdown-index
ENTRYPOINT ["/markdown-index"]
