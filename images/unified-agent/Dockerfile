FROM alpine:3.20.2

ARG UNIFIED_AGENT_VERSION=v24.3.2

RUN apk upgrade --no-cache; \
    apk add --no-cache \
    openjdk17-jre \
    ca-certificates  \
    curl \
    bash \
    && \
    update-ca-certificates

RUN curl -LO https://github.com/whitesource/unified-agent-distribution/releases/download/"${UNIFIED_AGENT_VERSION}"/wss-unified-agent.jar; \
    mkdir -p /wss && \
    mv wss-unified-agent.jar /wss/wss-unified-agent.jar

ENV LANG C.UTF-8

COPY start-wssagent.sh /
