FROM node:16.20.2-alpine AS node

FROM local/unified-agent:latest

COPY --from=node /usr/local/bin/node /usr/local/bin/node
COPY --from=node /usr/local/bin/npm /usr/local/bin/npm
COPY --from=node /usr/local/bin/yarn /usr/local/bin/yarn

COPY javascript-wss-unified-agent.config /wss