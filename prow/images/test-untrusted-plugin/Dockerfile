FROM golang:1.20.0-alpine3.16 AS builder

# Commit details

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . .

RUN apk add --no-cache bash upx dep git && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /test-untrusted -ldflags="-s -w" development/external-plugins/test-untrusted/cmd/main.go
FROM alpine:3.17.1

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit
COPY --from=builder /test-untrusted /external-plugin/
RUN apk add --no-cache ca-certificates bash && \
    chmod a+x /external-plugin/test-untrusted
WORKDIR /external-plugin
# for better access in a container
ENV PATH=$PATH:/external-plugin
ENTRYPOINT ["test-untrusted"]
