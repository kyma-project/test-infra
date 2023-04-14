FROM alpine:3.17.1 AS creds

SHELL ["/bin/ash", "option",  "-o", "pipefail"]
RUN apk add --no-cache curl && \
  curl -fsSL "https://github.com/GoogleCloudPlatform/docker-credential-gcr/releases/download/v2.1.5/docker-credential-gcr_linux_amd64-2.1.5.tar.gz" \
  | tar xz docker-credential-gcr \
  && chmod +x docker-credential-gcr && mv docker-credential-gcr /usr/bin/

FROM moby/buildkit:v0.11.1-rootless

COPY --from=creds /usr/bin/docker-credential-gcr /usr/bin/
RUN docker-credential-gcr configure-docker --registries=eu.gcr.io,europe-docker.pkg.dev

ENV USE_BUILDKIT=true

COPY ./image-builder /image-builder

ENTRYPOINT ["/image-builder"]