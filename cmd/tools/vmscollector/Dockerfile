FROM golang:1.23-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source to the Working Directory inside the container
COPY . .

WORKDIR /app/cmd/tools/vmscollector

# Build the Go app with static linking
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:3.20.3

LABEL io.kyma-project.source=github.com/kyma-project/test-infra/cmd/tools/vmscollector

# Copy the built Go app from the builder stage
COPY --from=builder /app/cmd/tools/vmscollector/main /vmscollector

RUN apk add --no-cache ca-certificates git && \
    chmod +x /vmscollector
ENTRYPOINT ["/vmscollector"]
