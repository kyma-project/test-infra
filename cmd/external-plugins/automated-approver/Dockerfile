FROM golang:1.24.2-alpine3.21 as builder

ARG BUILD_COMMIT_SHA="unknown"

RUN echo "Building commit SHA: ${BUILD_COMMIT_SHA}"

WORKDIR /app

COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source to the Working Directory inside the container
COPY . .

WORKDIR /app/cmd/external-plugins/automated-approver

# Build the Go app with static linking
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /automated-approver .

RUN apk add --no-cache ca-certificates git && \
    chmod +x /automated-approver

ENTRYPOINT ["/automated-approver"]