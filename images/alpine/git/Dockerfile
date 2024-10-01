ARG BASE_ALPINE_IMAGE

RUN echo "Cleaning image list format" && \
    export BASE_IMAGE=$(echo $BASE_ALPINE_IMAGE | sed 's/\[//g' | sed 's/\]//g') && \

FROM $BASE_IMAGE

RUN apk add --no-cache git git-daemon openssh

COPY github-known-hosts /github_known_hosts
COPY ssh-config /etc/ssh/ssh_config
