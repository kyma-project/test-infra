FROM alpine:3.16.2 AS creds

SHELL ["/bin/ash", "option",  "-o", "pipefail"]
RUN apk add --no-cache curl && \
    curl -fsSL "https://github.com/GoogleCloudPlatform/docker-credential-gcr/releases/download/v2.1.5/docker-credential-gcr_linux_amd64-2.1.5.tar.gz" \
  | tar xz docker-credential-gcr \
  && chmod +x docker-credential-gcr && mv docker-credential-gcr /usr/bin/

FROM moby/buildkit:v0.10.4-rootless

COPY --from=creds /usr/bin/docker-credential-gcr /usr/bin/
RUN docker-credential-gcr configure-docker

ENV USE_BUILDKIT=true

COPY ./image-builder /image-builder

ENTRYPOINT ["/image-builder"]