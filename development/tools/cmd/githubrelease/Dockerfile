FROM golang:1.19-alpine AS build

WORKDIR /build

COPY . .

RUN CGO_ENABLED=0 go build -o ./githubrelease ./development/tools/cmd/githubrelease && \
    chmod +x ./githubrelease

FROM node:18.11.0-alpine

LABEL source git@github.com:kyma-project/test-infra.git

WORKDIR /app

# Install dependencies
RUN apk update && \
    apk upgrade && \
    apk add --no-cache git openssh ca-certificates curl bash && \
    rm -rf /var/cache/apk/*

# Copy all needed files
RUN npm install -g lerna-changelog@0.8.0 && \
    npm install -g markdown-toc@1.2.0

ENV APP_PATH=/app

COPY ./changelog-generator/app/ /app/
COPY --from=build /build/githubrelease /githubrelease
