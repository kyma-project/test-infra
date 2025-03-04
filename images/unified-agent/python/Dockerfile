ARG BASE_UNIFIED_AGENT_IMAGE
FROM ${BASE_UNIFIED_AGENT_IMAGE}

### Install python3.x
RUN apk add --no-cache python3 py3-pip py3-virtualenv poetry

COPY images/unified-agent/python/python-wss-unified-agent.config /wss
# (2025-03-04)