FROM eu.gcr.io/kyma-project/test-infra/buildpack-java-node:v20200226-b9357193

# Commit details

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

# install unified agent
RUN curl -LJO https://github.com/whitesource/unified-agent-distribution/releases/download/v20.4.2.1/wss-unified-agent.jar
RUN mkdir -p /wss
RUN mv wss-unified-agent.jar /wss/wss-unified-agent.jar

# Prow Tools

COPY --from=eu.gcr.io/kyma-project/test-infra/prow-tools:v20200827-b59d2d67 /prow-tools /prow-tools
# for better access to prow-tools
ENV PATH=$PATH:/prow-tools