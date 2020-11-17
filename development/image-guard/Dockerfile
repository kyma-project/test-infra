FROM golang:1.15-alpine as builder

WORKDIR /build
COPY main.go .
COPY go.mod .
COPY go.sum .

RUN apk --no-cache add ca-certificates
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o image-guardd ./main.go

FROM scratch
ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

# copy image-guard binary
COPY --from=builder /build/image-guardd /

# add ca-certs to scratch image
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/image-guardd"]
