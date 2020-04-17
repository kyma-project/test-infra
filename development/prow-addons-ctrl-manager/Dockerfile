# Build the manager binary
FROM golang:1.13.10-alpine3.11 as builder

# Copy in the go src
WORKDIR /go/src/github.com/kyma-project/test-infra/development/prow-addons-ctrl-manager
COPY pkg/    pkg/
COPY cmd/    cmd/
COPY vendor/ vendor/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager github.com/kyma-project/test-infra/development/prow-addons-ctrl-manager/cmd/manager

# Copy the controller-manager into a thin image
FROM alpine:3.11
WORKDIR /

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

LABEL source=git@github.com:kyma-project/test-infra.git

COPY --from=builder /go/src/github.com/kyma-project/test-infra/development/prow-addons-ctrl-manager/manager .
ENTRYPOINT ["/manager"]
