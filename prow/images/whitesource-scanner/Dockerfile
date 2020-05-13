FROM eu.gcr.io/kyma-project/test-infra/buildpack-java-node:v20200226-b9357193

# install unified agent
RUN curl -LJO https://unified-agent.s3.amazonaws.com/wss-unified-agent-20.4.2.jar
RUN mkdir -p /wss
RUN mv wss-unified-agent-20.4.2.jar /wss/wss-unified-agent.jar
