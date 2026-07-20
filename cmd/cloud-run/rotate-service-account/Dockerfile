FROM europe-docker.pkg.dev/kyma-project/prod/external/library/golang:1.26.5-alpine3.23 as builder

WORKDIR /go/src/github.com/kyma-project/test-infra
COPY . .

RUN  CGO_ENABLED=0 go build -o /rotateserviceaccount -ldflags="-s -w" ./cmd/cloud-run/rotate-service-account

FROM europe-docker.pkg.dev/kyma-project/prod/external/library/alpine:3.24.1

COPY --from=builder /rotateserviceaccount /rotateserviceaccount
RUN apk add --no-cache ca-certificates && \
	chmod +x /rotateserviceaccount
ENTRYPOINT ["/rotateserviceaccount"]
