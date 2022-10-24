FROM golangci/golangci-lint:v1.50.1


# Commit details
ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

# Default to UTF-8 file.encoding
ENV LANG C.UTF-8

WORKDIR /workspace
COPY golangci-lint.sh /golangci-lint.sh
ENTRYPOINT ["/golangci-lint.sh"]
