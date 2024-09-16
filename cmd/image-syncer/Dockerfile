# Stage 1: Build the Go app
FROM golang:1.23-alpine as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the cmd/image-syncer directory to the Working Directory inside the container
COPY . .

WORKDIR /app/cmd/image-syncer

# Build the Go app with static linking
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Stage 2: Create the final image from scratch
FROM scratch

# Copy the built Go app from the builder stage
COPY --from=builder /app/cmd/image-syncer/main /image-syncer

# Command to run the executable
CMD ["/image-syncer"]