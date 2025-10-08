FROM golang:1.25.2-alpine3.22 as builder

WORKDIR /app

COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source to the Working Directory inside the container
COPY . .

WORKDIR /app/cmd/cloud-run/github-webhook-gateway

# Build the Go app with static linking
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
RUN apk add --no-cache ca-certificates

FROM scratch

LABEL io.kyma-project.source=github.com/kyma-project/test-infra/cmd/cloud-run/github-webhook-gateway

# Copy the built Go app from the builder stage
COPY --from=builder /app/cmd/cloud-run/github-webhook-gateway/main /github-webhook-gateway
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/github-webhook-gateway"]
