FROM golang:1.17.0-alpine3.14 AS builder

# Commit details

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . .

RUN apk --update --no-cache add bash upx dep git && \
    development/tools/build-cleaners.sh && \
    CGO_ENABLED=0 go build -o /prow-tools/jobguard -ldflags="-s -w" development/jobguard/cmd/jobguard/main.go && \
    CGO_ENABLED=0 go build -o /prow-tools/unique-jobs-name -ldflags="-s -w" development/checker/unique-jobs-name/main.go

FROM alpine:latest

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL IMAGE_COMMIT=$commit
RUN apk --update --no-cache add ca-certificates bash
COPY --from=builder /go/src/github.com/kyma-project/test-infra/development/tools/bin /prow-tools
COPY --from=builder /prow-tools/* /prow-tools/
WORKDIR /prow-tools
# for better access in a container
ENV PATH=$PATH:/prow-tools
