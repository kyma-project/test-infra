FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang:v20210804-d3a9d497

COPY sources.list /etc/apt/sources.list

RUN apt update && apt install libgit2-dev pkg-config -y

#TODO: remove it
RUN apt search libgit2-dev
