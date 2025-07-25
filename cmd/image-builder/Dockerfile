FROM golang:1.24.2-alpine3.21 AS builder
RUN apk add --no-cache ca-certificates curl

RUN curl -fsSL "https://github.com/GoogleCloudPlatform/docker-credential-gcr/releases/download/v2.1.22/docker-credential-gcr_linux_amd64-2.1.22.tar.gz" \
    | tar xz -C /usr/bin docker-credential-gcr && \
    chmod +x /usr/bin/docker-credential-gcr && \
    docker-credential-gcr configure-docker --registries=eu.gcr.io,europe-docker.pkg.dev

WORKDIR /app
COPY . .
RUN cd cmd/image-builder && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -buildvcs=false \
    -o /image-builder \
    -a \
    -ldflags '-w -s -extldflags "-static"' \
    .

FROM gcr.io/distroless/static:latest
WORKDIR /workspace
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/bin/docker-credential-gcr /usr/bin/
COPY --from=builder /root/.docker/config.json /home/nonroot/.docker/config.json
COPY --from=builder /image-builder /image-builder

ENTRYPOINT ["/image-builder"]
