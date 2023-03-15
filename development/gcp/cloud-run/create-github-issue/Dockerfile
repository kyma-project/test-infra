FROM golang:1.20.2-alpine3.17 as builder

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . .

RUN  CGO_ENABLED=0 go build -o /creategithubissue -ldflags="-s -w" ./development/gcp/cloud-run/create-github-issue

FROM alpine:3.17.2

COPY --from=builder /creategithubissue /creategithubissue
RUN apk add --no-cache ca-certificates && \
	chmod +x /creategithubissue
ENTRYPOINT ["/creategithubissue"]
