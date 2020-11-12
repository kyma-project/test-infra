# Basic node buildpack

FROM eu.gcr.io/kyma-project/prow/test-infra/bootstrap:v20181121-f3ea5ce

# Commit details

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

RUN curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add -
RUN echo "deb https://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list

RUN curl -sL https://deb.nodesource.com/setup_12.x | bash -
RUN apt-get update && apt-get install -y --no-install-recommends \
    libfontconfig \
    procps \
    nodejs \
    yarn \
    && apt-get clean

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

COPY --from=eu.gcr.io/kyma-project/test-infra/prow-tools:v20201110-cdbf6087 /prow-tools /prow-tools
# for better access to prow-tools
ENV PATH=$PATH:/prow-tools