FROM golang:1.19.2-alpine3.16 as builder

WORKDIR /go/src/github.com/kyma-project/test-infra
RUN ls -l
COPY . .

RUN ls -l
RUN  CGO_ENABLED=0 go build -o /updatesecretgithubchannel -ldflags="-s -w" ./development/gcp/cloud-run/updatesecretgithubchannel

FROM alpine:3.16.2

COPY --from=builder /updatesecretgithubchannel /updatesecretgithubchannel
RUN apk add --no-cache ca-certificates && \
	chmod +x /updatesecretgithubchannel
ENTRYPOINT ["/updatesecretgithubchannel"]
