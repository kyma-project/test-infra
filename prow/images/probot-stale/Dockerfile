FROM node:12-alpine

# Commit details

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

WORKDIR /app

RUN apk update && \
  apk upgrade && \
  apk add --no-cache git=2.24.4-r0

COPY package.json /app/package.json

# hadolint ignore=DL3016
RUN npm install . --no-optional

EXPOSE 3000

CMD ["npm", "run", "start"]
