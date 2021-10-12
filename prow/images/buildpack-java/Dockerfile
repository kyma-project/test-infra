# Basic node buildpack

FROM eu.gcr.io/kyma-project/test-infra/bootstrap:v20211005-0e8c8173

# Commit details

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

ENV JAVA_VERSION 8u212-b04
ENV JAVA_BASE_URL https://github.com/AdoptOpenJDK/openjdk8-binaries/releases/download/jdk8u212-b04/OpenJDK8U-jdk_
ENV JAVA_URL_VERSION 8u212b04

ENV JAVA_HOME /usr/local/openjdk-8
ENV PATH $JAVA_HOME/bin:$PATH

# Default to UTF-8 file.encoding
ENV LANG C.UTF-8

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates=20210119 p11-kit=0.23.22-1 maven=3.6.3-5 \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fLSs -o openjdk.tgz "${JAVA_BASE_URL}x64_linux_hotspot_${JAVA_URL_VERSION}.tar.gz"

RUN mkdir -p "$JAVA_HOME"; \
	tar --extract \
		--file openjdk.tgz \
		--directory "$JAVA_HOME" \
		--strip-components 1 \
		--no-same-owner \
	; \
	rm openjdk.tgz* \
	; \
    java -version

# Prow Tools
# hadolint ignore=DL3022
COPY --from=eu.gcr.io/kyma-project/test-infra/prow-tools:v20210401-294e46e5 /prow-tools /prow-tools
# for better access to prow-tools
ENV PATH=$PATH:/prow-tools
