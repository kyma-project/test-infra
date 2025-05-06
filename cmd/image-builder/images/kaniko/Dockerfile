FROM golang:1.24.2-alpine3.21 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
#RUN --mount=target=. \
RUN cd /app/cmd/image-builder && CGO_ENABLED=0 go build -buildvcs=false -o /image-builder -a -ldflags '-extldflags "-static"' .


FROM scratch

COPY --from=builder /image-builder /image-builder

ENTRYPOINT ["/image-builder"]