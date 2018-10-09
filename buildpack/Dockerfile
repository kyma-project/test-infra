FROM golang:1.11.0
LABEL source=https://github.com/kyma-project/test-infra.git

WORKDIR /

## NODE.JS & DOCKER DEPS
RUN curl -sL https://deb.nodesource.com/setup_10.x | bash -
RUN apt-get update && apt-get install -y --no-install-recommends \
    nodejs \ 
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg2 \
    software-properties-common \
    ssh-client \
    lsb-release && \
    apt-get autoremove -y && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* 

## GOLANG
# versions
ENV DEP_RELEASE_TAG v0.4.1
ENV GO_META_LINTER_RELEASE_TAG v2.0.8

# install dep
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

# install gometalinter
RUN curl https://raw.githubusercontent.com/alecthomas/gometalinter/master/scripts/install.sh | sh -s ${GO_META_LINTER_RELEASE_TAG}

# install go2xunit
RUN go get -u github.com/tebeka/go2xunit

# install mockery
RUN go get -u github.com/vektra/mockery/cmd/mockery

# install linters, prettier, and whitesource 
RUN npm install -g eslint-config-react-app \
	babel-eslint@^7.2.3 \
	eslint@^4.1.1 \
	eslint-plugin-flowtype@^2.34.1 \
	eslint-plugin-import@^2.6.0 \
	eslint-plugin-jsx-a11y@^5.1.1 \
	eslint-plugin-react@^7.1.0 \
	tslint@^5.11.0 \
	tslint-angular@^1.1.2 \
	tslint-config-prettier@^1.15.0 \
	typescript@^3.1.1 \
	prettier@^1.14.3 \
	whitesource@^18.9.1  

## GCLOUD
ENV CLOUD_SDK_VERSION=219.0.1
ENV PATH=/google-cloud-sdk/bin:/workspace:${PATH} \
    CLOUDSDK_CORE_DISABLE_PROMPTS=1

RUN wget -q https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${CLOUD_SDK_VERSION}-linux-x86_64.tar.gz && \
    tar xzf google-cloud-sdk-${CLOUD_SDK_VERSION}-linux-x86_64.tar.gz -C / && \
    rm google-cloud-sdk-${CLOUD_SDK_VERSION}-linux-x86_64.tar.gz && \
    /google-cloud-sdk/install.sh \
        --disable-installation-options \
        --bash-completion=false \
        --path-update=false \
        --usage-reporting=false && \
    gcloud components install alpha beta kubectl && \
    gcloud info 

# DOCKER
# Add the Docker apt-repository
RUN curl -fsSL https://download.docker.com/linux/$(. /etc/os-release; echo "$ID")/gpg \
    | apt-key add - && \
    add-apt-repository \
    "deb [arch=amd64] https://download.docker.com/linux/$(. /etc/os-release; echo "$ID") \
    $(lsb_release -cs) stable"

# Install Docker
RUN apt-get update && \
    apt-get install -y --no-install-recommends docker-ce=18.06.0* && \
    sed -i 's/cgroupfs_mount$/#cgroupfs_mount\n/' /etc/init.d/docker && \
    apt-get autoremove -y && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* 

# Move Docker's storage location
RUN echo 'DOCKER_OPTS="${DOCKER_OPTS} --data-root=/docker-graph"' | \
    tee --append /etc/default/docker
RUN mkdir /docker-graph

COPY ["start.sh", \
        "/usr/local/bin/"]

ENTRYPOINT [ "/usr/local/bin/start.sh" ]