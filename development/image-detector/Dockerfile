FROM golang:1.20.6-alpine3.18 as builder

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . .

RUN  CGO_ENABLED=0 go build -o /imagedetector -ldflags="-s -w" ./development/image-detector

FROM alpine:3.18.2

COPY --from=builder /imagedetector /imagedetector
RUN apk add --no-cache ca-certificates git && \
	chmod +x /imagedetector
ENTRYPOINT ["/imagedetector"]
