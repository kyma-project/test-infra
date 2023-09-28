FROM local/alpine:latest AS download
ARG HADOLINT_VERSION=v2.12.0
RUN apk add --no-cache curl; \
    curl -Lo /bin/hadolint "https://github.com/hadolint/hadolint/releases/download/${HADOLINT_VERSION}/hadolint-Linux-$(uname -m)" && chmod +x /bin/hadolint

FROM local/alpine:latest

COPY --from=download /bin/hadolint /bin/hadolint

ENTRYPOINT [ "/bin/hadolint" ]