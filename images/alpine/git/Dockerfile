FROM local/alpine:latest

RUN apk add --no-cache git git-daemon openssh

COPY github-known-hosts /github_known_hosts
COPY ssh-config /etc/ssh/ssh_config