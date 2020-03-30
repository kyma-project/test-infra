FROM eu.gcr.io/kyma-project/test-infra/buildpack-java-node:v20200226-b9357193

# install unified agent
RUN curl -LJO https://github.com/whitesource/unified-agent-distribution/releases/download/v19.6.1/wss-unified-agent.jar
RUN mkdir -p /wss
RUN mv wss-unified-agent.jar /wss/wss-unified-agent.jar