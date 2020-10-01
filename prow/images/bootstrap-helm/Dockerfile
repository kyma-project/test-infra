FROM eu.gcr.io/kyma-project/test-infra/bootstrap:v20191120-ef8b2357

# Commit details

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

ENV HELM_VERSION="v3.2.1"

RUN wget https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz -O - | tar -xzO linux-amd64/helm > /usr/local/bin/helm &&\
	  chmod +x /usr/local/bin/helm &&\
      rm -rf helm-${HELM_VERSION}-linux-amd64.tar.gz linux-amd64

# Prow Tools

COPY --from=eu.gcr.io/kyma-project/test-infra/prow-tools:v20200827-b59d2d67 /prow-tools /prow-tools
# for better access to prow-tools
ENV PATH=$PATH:/prow-tools