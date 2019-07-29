FROM eu.gcr.io/kyma-project/test-infra/buildpack-java:v20190729-5052901

# install unified agent
RUN curl -LJO https://github.com/whitesource/unified-agent-distribution/releases/download/v19.6.1/wss-unified-agent.jar
RUN mkdir -p /wss
RUN mv wss-unified-agent.jar /wss/wss-unified-agent.jar

# add config file
COPY wss-unified-agent.config /wss/wss-unified-agent.config