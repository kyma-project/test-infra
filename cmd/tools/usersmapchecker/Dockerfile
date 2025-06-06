FROM golang:1.24.2-alpine3.21 as builder

WORKDIR /app

COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source to the Working Directory inside the container
COPY . .

WORKDIR /app/cmd/tools/usersmapchecker

# Build the Go app with static linking
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:3.20.3

LABEL io.kyma-project.source=github.com/kyma-project/test-infra/cmd/tools/usersmapchecker

# Copy the built Go app from the builder stage
COPY --from=builder /app/cmd/tools/usersmapchecker/main /usersmapchecker

RUN apk add --no-cache ca-certificates git && \
    chmod +x /usersmapchecker
ENTRYPOINT ["/usersmapchecker"]
