FROM golang:1.11.0
LABEL source=https://github.com/kyma-project/test-infra.git

WORKDIR /

## NODE.JS
RUN curl -sL https://deb.nodesource.com/setup_10.x | bash -
RUN apt-get update && apt-get install -y nodejs uidmap libseccomp-dev

RUN nodejs -v

## GOLANG
# install dep
RUN go get -u github.com/golang/dep/cmd/dep

# install gometalinter
RUN go get -u github.com/alecthomas/gometalinter \
        && gometalinter --install

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
	tslint \
	tslint-angular \
	tslint-config-prettier \
	typescript \
	prettier \
	whitesource  

## Install gcloud

ENV PATH=/google-cloud-sdk/bin:/workspace:${PATH} \
    CLOUDSDK_CORE_DISABLE_PROMPTS=1

RUN wget -q https://dl.google.com/dl/cloudsdk/channels/rapid/google-cloud-sdk.tar.gz && \
    tar xzf google-cloud-sdk.tar.gz -C / && \
    rm google-cloud-sdk.tar.gz && \
    /google-cloud-sdk/install.sh \
        --disable-installation-options \
        --bash-completion=false \
        --path-update=false \
        --usage-reporting=false && \
    gcloud components install alpha beta kubectl && \
    gcloud info 

#
# BEGIN: DOCKER IN DOCKER SETUP
#

# Install Docker deps
RUN apt-get update && apt-get install -y --no-install-recommends \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg2 \
    software-properties-common \
    lsb-release

# Add the Docker apt-repository
RUN curl -fsSL https://download.docker.com/linux/$(. /etc/os-release; echo "$ID")/gpg \
    | apt-key add - && \
    add-apt-repository \
    "deb [arch=amd64] https://download.docker.com/linux/$(. /etc/os-release; echo "$ID") \
    $(lsb_release -cs) stable"

# Install Docker
RUN apt-get update && \
    apt-get install -y --no-install-recommends docker-ce=18.06.0* && \
    sed -i 's/cgroupfs_mount$/#cgroupfs_mount\n/' /etc/init.d/docker

# Move Docker's storage location
RUN echo 'DOCKER_OPTS="${DOCKER_OPTS} --data-root=/docker-graph"' | \
    tee --append /etc/default/docker
RUN mkdir /docker-graph

#
# END: DOCKER IN DOCKER SETUP
#

COPY ["start.sh", \
        "/usr/local/bin/"]

ENTRYPOINT [ "/usr/local/bin/start.sh" ]