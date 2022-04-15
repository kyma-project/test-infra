FROM golang:1.18.0-alpine3.15 AS builder

# Commit details

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

WORKDIR /usr/local/go/src/github.com/kyma-project/test-infra
COPY . .
RUN apk add --no-cache bash dep git && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /automerge-notification -ldflags="-s -w" ./development/external-plugins/automerge-notification/main.go \

FROM alpine:3.15.4

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit
# hadolint ignore=DL3023
COPY --from=builder /automerge-notification /external-plugin/
RUN apk add --no-cache ca-certificates git && \
	chmod a+x /external-plugin/automerge-notification
WORKDIR /external-plugin
# for better access in a container
ENV PATH=$PATH:/external-plugin
ENTRYPOINT ["automerge-notification"]
