FROM alpine:edge

RUN apk add --no-cache buildah bash yq jq

COPY builder.sh /builder.sh

ENTRYPOINT ["/builder.sh"]
