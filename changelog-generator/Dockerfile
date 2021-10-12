FROM node:10.11.0-alpine

LABEL source git@github.com:kyma-project/test-infra.git

WORKDIR /app

# Install dependencies
RUN apk update && \
    apk upgrade && \
    apk add --no-cache git=2.18.4-r0 openssh=7.7_p1-r4 && \
    rm -rf /var/cache/apk/*

# Copy all needed files
RUN npm install -g lerna-changelog@0.8.0 && \
    npm install -g markdown-toc@1.2.0

ENV APP_PATH=/app

COPY /app/ /app/
