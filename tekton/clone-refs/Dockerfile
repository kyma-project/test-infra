FROM golang:1.20.2-alpine3.17 as builder

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . .

RUN  CGO_ENABLED=0 go build -o /clonerefs -ldflags="-s -w" ./tekton/clonerefs

FROM alpine:3.17.2

COPY --from=builder /clonerefs /clonerefs
RUN apk add --no-cache ca-certificates git && \
	chmod +x /clonerefs
ENTRYPOINT ["/clonerefs"]
