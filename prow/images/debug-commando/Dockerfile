# Image with tools for troubleshooting and debugging.

FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.17.0

# Commit details

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

# Install additional tools

RUN apt-get update && apt-get install -y --no-install-recommends \
    rsync \
    procps \
    && apt-get clean

# Install python

RUN apt-get update && apt-get install -y \
	python-setuptools \
	python-pip \
    && yes | pip install cgroup-utils \
    && apt-get clean

RUN GO111MODULE="on" go install golang.org/x/tools/cmd/goimports  \
    github.com/hashicorp/hcl/hcl/printer \
    golang.org/x/lint/golint \
    github.com/vektra/mockery \
    github.com/ericchiang/pup \
    github.com/kisielk/errcheck

COPY ./license-puller.sh /license-puller.sh
ENV LICENSE_PULLER_PATH=/license-puller.sh

# Prow Tools

COPY --from=eu.gcr.io/kyma-project/test-infra/prow-tools:v20210331-4a336452 /prow-tools /prow-tools
# for better access to prow-tools
ENV PATH=$PATH:/prow-tools
