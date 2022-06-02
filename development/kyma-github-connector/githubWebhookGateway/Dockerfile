FROM golang:1.18.3-alpine3.15 as builder

WORKDIR /build

COPY cmd cmd/
COPY pkg pkg/
COPY go.mod .
COPY go.sum .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o ghwebhookgateway ./cmd/main.go

FROM scratch
ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

# copy ghwebhookgateway binary
COPY --from=builder /build/ghwebhookgateway /

# add ca-certs to scratch image
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/ghwebhookgateway"]
