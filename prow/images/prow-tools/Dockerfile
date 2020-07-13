ARG commit

FROM golang:1.14.2-alpine3.11 AS builder

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . .

RUN apk --update --no-cache add bash upx dep git && \
    development/tools/build-cleaners.sh && \
    cd development/jobguard/ && CGO_ENABLED=0 go build -o /prow-tools/jobguard -ldflags="-s -w" cmd/jobguard/main.go && \
    cd ../checker/ && CGO_ENABLED=0 go build -o /prow-tools/unique-jobs-name -ldflags="-s -w" unique-jobs-name/main.go

ENV GOBIN=/prow-tools
RUN cd development/checker && go get k8s.io/test-infra/prow/cmd/checkconfig@v0.0.0-20200320172837-fbc86f22b087

FROM alpine:latest

ENV IMAGE_COMMIT=$commit
LABEL image_commit=$commit
RUN apk --update --no-cache add ca-certificates bash
COPY --from=builder /go/src/github.com/kyma-project/test-infra/development/tools/bin /prow-tools
COPY --from=builder /prow-tools/* /prow-tools/
WORKDIR /prow-tools
# for better access in a container
ENV PATH=$PATH:/prow-tools
