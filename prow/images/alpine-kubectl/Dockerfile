FROM alpine:3.11

LABEL source=git@github.com:kyma-project/test-infra.git

ENV HELM_VERSION="v2.16.1"
ENV KUBECTL_VERSION="v1.17.3"

RUN apk update &&\
    apk add --no-cache openssl coreutils curl bash jq &&\
	wget https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl -O /usr/local/bin/kubectl &&\
	chmod +x /usr/local/bin/kubectl &&\
	wget https://storage.googleapis.com/kubernetes-helm/helm-${HELM_VERSION}-linux-amd64.tar.gz -O - | tar -xzO linux-amd64/helm > /usr/local/bin/helm &&\
	chmod +x /usr/local/bin/helm &&\
	helm init --client-only

ENTRYPOINT ["/bin/bash"]
