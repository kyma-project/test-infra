FROM golang:1.22.3-alpine AS build

WORKDIR /build

COPY . .

RUN CGO_ENABLED=0 go build -o ./githubrelease ./cmd/tools/githubrelease && \
    chmod +x ./githubrelease

FROM node:22.1.0-alpine

LABEL source git@github.com:kyma-project/test-infra.git

WORKDIR /app

# Install dependencies
RUN apk update && \
    apk upgrade && \
    apk add --no-cache git openssh ca-certificates curl bash && \
    rm -rf /var/cache/apk/*

# Copy all needed files
RUN npm install -g lerna-changelog@2.2.0 && \
    npm install -g markdown-toc@1.2.0

ENV APP_PATH=/app

COPY ./experimental/changelog-generator/app/ /app/
COPY --from=build /build/githubrelease /githubrelease
