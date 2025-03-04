ARG BASE_ALPINE_IMAGE
FROM ${BASE_ALPINE_IMAGE}

RUN apk add --no-cache git git-daemon openssh

COPY images/alpine/git/github-known-hosts /github_known_hosts
COPY images/alpine/git/ssh-config /etc/ssh/ssh_config
# (2025-03-04)