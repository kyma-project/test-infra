FROM alpine:3.18.2 AS base

RUN apk add --no-cache curl
RUN curl -Lo /bin/hadolint https://github.com/hadolint/hadolint/releases/download/v2.12.0/hadolint-Linux-x86_64 && chmod +x /bin/hadolint
#TODO: Split this image into 2 seperate ones, one for hadolint and one for shellcheck. Probably tree structure would be the best.
FROM alpine:3.18.2
COPY --from=base /bin/hadolint /bin/hadolint
RUN apk add --no-cache bash shellcheck

ENTRYPOINT ["/bin/hadolint"]