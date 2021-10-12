FROM golang:1.17.0-alpine3.14 AS builder

# Commit details

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . .

RUN apk add --no-cache 'bash>=5.1' 'upx>=3.96' 'dep>=0.5' 'git>=2.32' && \
    development/tools/build-cleaners.sh && \
    CGO_ENABLED=0 go build -o /prow-tools/jobguard -ldflags="-s -w" development/jobguard/cmd/jobguard/main.go && \
    CGO_ENABLED=0 go build -o /prow-tools/unique-jobs-name -ldflags="-s -w" development/checker/unique-jobs-name/main.go

FROM alpine:3.13

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit
RUN apk add --no-cache 'ca-certificates>=20191127' 'bash>=5.1'
COPY --from=builder /go/src/github.com/kyma-project/test-infra/development/tools/bin /prow-tools
COPY --from=builder /prow-tools/* /prow-tools/
WORKDIR /prow-tools
# for better access in a container
ENV PATH=$PATH:/prow-tools
