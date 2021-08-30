FROM golang:1.17.0-alpine

RUN apk add --no-cache build-base musl-dev containers-common crun device-mapper fuse-overlayfs gpgme-dev libseccomp-dev shadow-uidmap slirp4netns btrfs-progs-dev lvm2-dev
#RUN addgroup -g 1000 user && \
#    adduser -S -s /bin/ash -u 1000 -G user user && \
#    echo user:10000:65536 >> /etc/subuid && \
#    echo user:10000:65536 >> /etc/subgid
#USER user
