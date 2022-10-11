# Basic node buildpack

FROM eu.gcr.io/kyma-project/test-infra/bootstrap:v20221011-310d4b98

# Commit details

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add -
RUN echo "deb https://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list

RUN curl -sL https://deb.nodesource.com/setup_12.x | bash -
RUN apt-get update && apt-get install -y --no-install-recommends \
    libfontconfig1 \
    procps \
    nodejs \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

RUN npm install -g eslint-config-react-app@^3.0.4 \
    babel-eslint@^9.0.0 \
    eslint@^5.6.1 \
    eslint-plugin-flowtype@^2.50.3 \
    eslint-plugin-import@^2.14.0 \
    eslint-plugin-jsx-a11y@^6.1.2 \
    eslint-plugin-react@^7.11.1 \
    tslint@^5.11.0 \
    tslint-angular@^1.1.2 \
    tslint-config-prettier@^1.15.0 \
    typescript@^3.1.3 \
    prettier@^1.14.3

COPY ./license-puller.sh /license-puller.sh
ENV LICENSE_PULLER_PATH=/license-puller.sh

# Prow Tools
# hadolint ignore=DL3022
COPY --from=eu.gcr.io/kyma-project/test-infra/prow-tools:v20210401-294e46e5 /prow-tools /prow-tools
# for better access to prow-tools
ENV PATH=$PATH:/prow-tools
