FROM eu.gcr.io/kyma-project/test-infra/bootstrap:v20191120-ef8b2357

ENV HELM_VERSION="dev-v3"

RUN wget -q https://storage.googleapis.com/kubernetes-helm/helm-${HELM_VERSION}-linux-amd64.tar.gz -O - | tar -xzO linux-amd64/helm > /usr/local/bin/helm \
    && chmod +x /usr/local/bin/helm

RUN helm init --client-only
