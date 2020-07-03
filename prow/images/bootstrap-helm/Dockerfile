FROM eu.gcr.io/kyma-project/test-infra/bootstrap:v20191120-ef8b2357

ENV HELM_VERSION="v3.2.1"

RUN wget https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz -O - | tar -xzO linux-amd64/helm > /usr/local/bin/helm &&\
	  chmod +x /usr/local/bin/helm &&\
      rm -rf helm-${HELM_VERSION}-linux-amd64.tar.gz linux-amd64

