FROM europe-docker.pkg.dev/kyma-project/prod/testimages/buildpack-go:v20240320-a2016d2a AS builder

WORKDIR /
COPY . /app/
RUN cd /app/cmd/image-builder && CGO_ENABLED=0 go build -buildvcs=false -o /app/image-builder -a -ldflags '-extldflags "-static"' .


FROM gcr.io/kaniko-project/executor:v1.21.1

COPY --from=builder /app/image-builder /image-builder

ENTRYPOINT ["/image-builder"]
