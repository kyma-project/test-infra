# Stage 1: Build the Go app
FROM golang:1.23.1-alpine3.20 as builder

# Install certificate authorities
RUN apk update && apk add --no-cache ca-certificates

WORKDIR /app

COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source to the Working Directory inside the container
COPY . .

WORKDIR /app/cmd/image-syncer

# Build the Go app with static linking
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Stage 2: Create the final image from scratch
FROM scratch

LABEL io.kyma-project.source=github.com/kyma-project/test-infra/cmd/image-syncer

# Copy the built Go app from the builder stage
COPY --from=builder /app/cmd/image-syncer/main /image-syncer

# Copy CA certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Command to run the executable
ENTRYPOINT ["/image-syncer"]