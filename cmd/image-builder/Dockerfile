FROM golang:1.24.2-alpine3.21 AS builder
WORKDIR /
COPY . /app/
RUN cd /app/cmd/image-builder && CGO_ENABLED=0 go build -buildvcs=false -o /app/image-builder -a -ldflags '-extldflags "-static"' .

FROM scratch
COPY --from=builder /app/image-builder /image-builder
WORKDIR /workspace
ENTRYPOINT ["/image-builder"]
