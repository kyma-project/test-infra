FROM moby/buildkit:v0.10.4-rootless

ENV USE_BUILDKIT=true

COPY ./kaniko-build /kaniko-build

ENTRYPOINT ["/kaniko-build"]