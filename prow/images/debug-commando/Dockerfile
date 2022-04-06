# Image with tools for troubleshooting and debugging.

FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.18

# Commit details

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

# Install additional tools

RUN apt-get update && apt-get install -y --no-install-recommends \
    rsync \
    procps \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install python
## python 2 pip no longer supported
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN apt-get update && apt-get install -y --no-install-recommends \
	python-setuptools \
	python3-pip \
    && yes | pip install --no-cache-dir cgroup-utils==0.8 \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

RUN GO111MODULE="on" go install golang.org/x/tools/cmd/goimports  \
    github.com/hashicorp/hcl/hcl/printer \
    golang.org/x/lint/golint \
    github.com/vektra/mockery \
    github.com/ericchiang/pup \
    github.com/kisielk/errcheck

COPY ./license-puller.sh /license-puller.sh
ENV LICENSE_PULLER_PATH=/license-puller.sh

# Prow Tools
# hadolint ignore=DL3022
COPY --from=eu.gcr.io/kyma-project/test-infra/prow-tools:v20210331-4a336452 /prow-tools /prow-tools
# for better access to prow-tools
ENV PATH=$PATH:/prow-tools
