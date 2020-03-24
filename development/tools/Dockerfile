FROM golang:1.12.10-alpine3.10 as builder

RUN apk --update --no-cache add bash upx
ENV SRC_DIR=/go/src/github.com/kyma-project/test-infra/development/tools
ADD . $SRC_DIR

WORKDIR $SRC_DIR

RUN ./build-all.sh

FROM alpine:latest

RUN apk --update --no-cache add ca-certificates bash
COPY --from=builder /go/src/github.com/kyma-project/test-infra/development/tools/bin/ /usr/local/bin/
