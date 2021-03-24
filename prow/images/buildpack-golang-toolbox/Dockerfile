# Basic golang buildpack

FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.15.7

# Commit details

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

ARG linterVersion
ENV LINTER_VERSION=$linterVersion

# Install additional tools

RUN apt-get update && apt-get install -y --no-install-recommends \
    rsync=3.1.2-1+deb9u2 \
    procps=2:3.3.12-3+deb9u1 \
    && apt-get clean

RUN go get golang.org/x/tools/cmd/goimports \
    golang.org/x/lint/golint \
    github.com/ericchiang/pup \
    github.com/kisielk/errcheck && \
    GO111MODULE="on" go get github.com/vektra/mockery && \
    GO111MODULE="on" go get sigs.k8s.io/kind@v0.5.1 && \
    GO111MODULE="off" go get -u github.com/maxbrunsfeld/counterfeiter

COPY ./license-puller.sh /license-puller.sh
ENV LICENSE_PULLER_PATH=/license-puller.sh

# Install linter

RUN curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | bash -s -- -b ${GOPATH}/bin ${LINTER_VERSION}

# Prow Tools

COPY --from=eu.gcr.io/kyma-project/test-infra/prow-tools:v20201208-c200b51a /prow-tools /prow-tools
# for better access to prow-tools
ENV PATH=$PATH:/prow-tools
