FROM --platform=$BUILDPLATFORM golang:1.24.2-alpine3.21 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download -x

#COPY . .
ARG TARGETOS TARGETARCH
RUN --mount=target=. --mount=type=cache,target=/root/.cache/go-build cd /app/cmd/image-builder && CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -buildvcs=false -o /image-builder -a -ldflags '-extldflags "-static"' .


FROM scratch

COPY --from=builder /image-builder /image-builder

ENTRYPOINT ["/image-builder"]