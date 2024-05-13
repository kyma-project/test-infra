FROM golangci/golangci-lint:v1.58.1


# Commit details
ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

# Default to UTF-8 file.encoding
ENV LANG C.UTF-8

WORKDIR /workspace
COPY golangci-lint.sh /golangci-lint.sh

# install libgit2 for serverless components
RUN apt-get update && apt-get install --no-install-recommends --yes libgit2-dev \
  && apt-get clean && rm -rf /var/lib/apt/lists/*

ENTRYPOINT ["/golangci-lint.sh"]
