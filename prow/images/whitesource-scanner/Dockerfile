FROM eu.gcr.io/kyma-project/test-infra/buildpack-java:v20190116-69aa0a1

# install unified agent
RUN curl -LJO https://github.com/whitesource/unified-agent-distribution/releases/download/v19.6.1/wss-unified-agent.jar

# add config file
COPY wss-unified-agent.config wss-unified-agent.config