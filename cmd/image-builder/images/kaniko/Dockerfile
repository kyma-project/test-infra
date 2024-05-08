FROM europe-docker.pkg.dev/kyma-project/prod/testimages/buildpack-go:v20240409-9277fe7f AS builder

WORKDIR /
COPY . /app/
RUN cd /app/cmd/image-builder && CGO_ENABLED=0 go build -buildvcs=false -o /app/image-builder -a -ldflags '-extldflags "-static"' .


FROM ubuntu:latest

COPY --from=builder /app/image-builder /image-builder

RUN apt-get update && apt-get install -y tree
RUN tree

ENTRYPOINT ["/image-builder"]
