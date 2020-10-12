FROM golang:1.15-alpine as builder

WORKDIR /build
COPY main.go .
COPY go.mod .
COPY go.sum .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o image-guardd ./main.go

FROM scratch

COPY --from=builder /build/image-guardd /

ENTRYPOINT ["/image-guardd"]
