FROM alpine:3.10

LABEL source=git@github.com:kyma-project/test-infra.git

ENV HELM_VERSION="v2.16.1"
ENV KUBECTL_VERSION="v1.16.3"

RUN apk update &&\
	apk add openssl coreutils curl bash &&\
	curl -LO curl -LO https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl &&\
	chmod +x ./kubectl &&\
	mv ./kubectl /usr/local/bin/kubectl &&\
	wget -q https://storage.googleapis.com/kubernetes-helm/helm-${HELM_VERSION}-linux-amd64.tar.gz -O - | tar -xzO linux-amd64/helm > /usr/local/bin/helm &&\
	chmod +x /usr/local/bin/helm &&\
	helm init --client-only

ENTRYPOINT ["/bin/bash"]
