FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang:v20190116-69aa0a1

# install node & npm
RUN curl -sL https://deb.nodesource.com/setup_8.x | bash -
RUN apt-get install -y nodejs

# install snyk CLI app
RUN npm install -g snyk@1.134.2
