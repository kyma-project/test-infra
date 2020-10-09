FROM eu.gcr.io/kyma-project/prow/bootstrap:0.0.1

# Commit details

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

COPY cleaner.sh cleaner.sh
ENTRYPOINT ["./cleaner.sh"]

# Prow Tools

COPY --from=eu.gcr.io/kyma-project/test-infra/prow-tools:v20201008-a2fd337c /prow-tools /prow-tools
# for better access to prow-tools
ENV PATH=$PATH:/prow-tools