FROM eu.gcr.io/kyma-project/test-infra/bootstrap:v20221121-4ded4f32

# Commit details

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

WORKDIR /workspace
COPY cleaner.sh cleaner.sh
ENTRYPOINT ["./cleaner.sh"]

# Prow Tools
# hadolint ignore=DL3022
COPY --from=eu.gcr.io/kyma-project/test-infra/prow-tools:v20221118-5c35cc55 /prow-tools /prow-tools
# for better access to prow-tools
ENV PATH=$PATH:/prow-tools
