#FROM golang:1.17.0-alpine
#
#RUN apk add --no-cache build-base musl-dev containers-common crun device-mapper fuse-overlayfs gpgme-dev libseccomp-dev shadow-uidmap slirp4netns btrfs-progs-dev lvm2-dev

FROM alpine:edge

RUN apk add --no-cache buildah bash yq jq git fuse-overlayfs shadow slirp4netns
RUN sed -i 's/^#mount_program/mount_program/' /etc/containers/storage.conf && \
    addgroup -g 1000 builder && \
    adduser -S -s /bin/ash -u 1000 -G builder builder && \
    usermod --add-subuids 100000-165535 builder && \
    usermod --add-subgids 100000-165535 builder
#USER builder
COPY image-builder.sh /image-builder.sh

ENTRYPOINT ["/image-builder.sh"]
