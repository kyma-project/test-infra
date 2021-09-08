#FROM golang:1.17.0-alpine
#
#RUN apk add --no-cache build-base musl-dev containers-common crun device-mapper fuse-overlayfs gpgme-dev libseccomp-dev shadow-uidmap slirp4netns btrfs-progs-dev lvm2-dev

FROM alpine:edge

RUN apk add --no-cache buildah bash yq jq git
RUN addgroup -g 1000 builder && \
    adduser -S -s /bin/ash -u 1000 -G builder builder && \
    echo builder:10000:65536 >> /etc/subuid && \
    echo builder:10000:65536 >> /etc/subgid && \
    chmod 4755 /usr/bin/newgidmap && \
    chmod 4755 /usr/bin/newuidmap
USER builder
COPY image-builder.sh /image-builder.sh

ENTRYPOINT ["/image-builder.sh"]
