# Node buildpack with chromium and its dependencies

FROM eu.gcr.io/kyma-project/prow/test-infra/buildpack-node:v20181130-b28250b

ENV PUPPETEER_EXECUTABLE_PATH /usr/bin/chromium
ENV PUPPETEER_SKIP_CHROMIUM_DOWNLOAD true

RUN apt-get update && apt-get install -y --no-install-recommends \
    procps=2:3.3.12-3+deb9u1 \
    gconf-service=3.2.6-4+b1 \
    libasound2=1.1.3-5 \
    libatk1.0-0=2.22.0-1 \
    libatk-bridge2.0-0=2.22.0-2 \
    libc6=2.24-11+deb9u4 \
    libcairo2=1.14.8-1+deb9u1 \
    libcups2=2.2.1-8+deb9u6 \
    libdbus-1-3=1.10.32-0+deb9u1 \
    libexpat1=2.2.0-2+deb9u3 \
    libfontconfig1=2.11.0-6.7+b1 \
    libgcc1=1:6.3.0-18+deb9u1 \
    libgconf-2-4=3.2.6-4+b1 \
    libgdk-pixbuf2.0-0=2.36.5-2+deb9u2 \
    libglib2.0-0=2.50.3-2+deb9u2 \
    libgtk-3-0=3.22.11-1 \
    libnspr4=2:4.12-6 \
    libpango-1.0-0=1.40.5-1 \
    libpangocairo-1.0-0=1.40.5-1 \
    libstdc++6=6.3.0-18+deb9u1 \
    libx11-6=2:1.6.4-3+deb9u4 \
    libx11-xcb1=2:1.6.4-3+deb9u4 \
    libxcb1=1.12-1 \
    libxcomposite1=1:0.4.4-2 \
    libxcursor1=1:1.1.14-1+deb9u2 \
    libxdamage1=1:1.1.4-2+b3 \
    libxext6=2:1.3.3-1+b2 \
    libxfixes3=1:5.0.3-1 \
    libxi6=2:1.7.9-1 \
    libxrandr2=2:1.5.1-1 \
    libxrender1=1:0.9.10-1 \
    libxss1=1:1.2.2-1 \
    libxtst6=2:1.2.3-1 \
    ca-certificates=20200601~deb9u2 \
    fonts-liberation=1:1.07.4-2 \
    libappindicator1=0.4.92-4 \
    libnss3=2:3.26.2-1.1+deb9u2 \
    lsb-release=9.20161125 \
    xdg-utils=1.1.1-1+deb9u2 \
    wget=1.18-5+deb9u3 \
    chromium=73.0.3683.75-1~deb9u1 \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*
