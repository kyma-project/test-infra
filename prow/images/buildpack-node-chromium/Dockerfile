# Node buildpack with chromium and its dependencies

FROM eu.gcr.io/kyma-project/prow/test-infra/buildpack-node:v20181119-afd3fbd

RUN curl -sL https://deb.nodesource.com/setup_10.x | bash -
RUN apt-get update && apt-get install -y --no-install-recommends \
    procps \
    gconf-service \
    libasound2 \
    libatk1.0-0 \
    libatk-bridge2.0-0 \
    libc6 \
    libcairo2 \
    libcups2 \
    libdbus-1-3 \
    libexpat1 \
    libfontconfig1 \
    libgcc1 \
    libgconf-2-4 \
    libgdk-pixbuf2.0-0 \
    libglib2.0-0 \
    libgtk-3-0 \
    libnspr4 \
    libpango-1.0-0 \
    libpangocairo-1.0-0 \
    libstdc++6 \
    libx11-6 \
    libx11-xcb1 \
    libxcb1 \
    libxcomposite1 \
    libxcursor1 \
    libxdamage1 \
    libxext6 \
    libxfixes3 \
    libxi6 \
    libxrandr2 \
    libxrender1 \
    libxss1 \
    libxtst6 \
    ca-certificates \
    fonts-liberation \
    libappindicator1 \
    libnss3 \
    lsb-release \
    xdg-utils \
    wget \
    chromium=70.0.3538.110-1~deb9u1 \
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
    prettier@^1.14.3 \
    whitesource@^18.10.1

ENV PUPPETEER_EXECUTABLE_PATH /usr/bin/chromium
ENV PUPPETEER_SKIP_CHROMIUM_DOWNLOAD true

