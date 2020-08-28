FROM node:12-alpine

# Commit details

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

WORKDIR /app

RUN apk update && \
  apk upgrade && \
  apk add git

COPY package.json /app/package.json

RUN npm install . --no-optional

EXPOSE 3000

CMD ["npm", "run", "start"]
