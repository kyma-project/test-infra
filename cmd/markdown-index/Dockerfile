FROM golang:1.25.1-alpine3.22 as builder

WORKDIR /app

COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source to the Working Directory inside the container
COPY . .

WORKDIR /app/cmd/markdown-index

# Build the Go app with static linking
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .



FROM alpine:3.22.1 as runtime

LABEL io.kyma-project.source=github.com/kyma-project/test-infra/cmd/markdown-index

# Copy the built Go app from the builder stage
COPY --from=builder /app/cmd/markdown-index/main /markdown-index

RUN apk add --no-cache ca-certificates git && \
    chmod +x /markdown-index
RUN echo "Dockerfile build reached this step"
ENTRYPOINT ["/markdown-index"]


FROM alpine:3.22.1 as prod

LABEL io.kyma-project.source=github.com/kyma-project/test-infra/cmd/markdown-index

# Copy the built Go app from the builder stage
COPY --from=builder /app/cmd/markdown-index/main /markdown-index

RUN apk add --no-cache ca-certificates git && \
    chmod +x /markdown-index

RUN echo "Dockerfile build didnt reach this step"
ENTRYPOINT ["/markdown-index"]
