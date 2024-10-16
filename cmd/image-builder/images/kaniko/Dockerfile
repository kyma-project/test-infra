FROM europe-docker.pkg.dev/kyma-project/prod/buildpack-go:v20241003-9d1845c9 AS builder

WORKDIR /
COPY . /app/
RUN cd /app/cmd/image-builder && CGO_ENABLED=0 go build -buildvcs=false -o /app/image-builder -a -ldflags '-extldflags "-static"' .


FROM gcr.io/kaniko-project/executor:v1.23.2

COPY --from=builder /app/image-builder /image-builder

ENTRYPOINT ["/image-builder"]