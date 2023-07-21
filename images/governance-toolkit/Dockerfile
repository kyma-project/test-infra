FROM alpine:latest AS base

RUN apk add --no-cache curl
RUN curl -Lo /bin/hadolint https://github.com/hadolint/hadolint/releases/download/v2.12.0/hadolint-Linux-x86_64 && chmod +x /bin/hadolint

FROM alpine:latest
COPY --from=base /bin/hadolint /bin/hadolint
RUN apk add --no-cache bash shellcheck

ENTRYPOINT ["/bin/hadolint"]