FROM hashicorp/terraform:1.3.7

# Commit details
ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit


RUN apk add --no-cache unzip && \
    mkdir /tfplugins && \
    curl -fLSs -o /terraform-provider-google_4.50.0.zip "https://releases.hashicorp.com/terraform-provider-google/4.50.0/terraform-provider-google_4.50.0_linux_amd64.zip" && \
    unzip -q /terraform-provider-google_4.50.0.zip -d /tfplugins//terraform-provider-google_4.50.0


WORKDIR $GOPATH
ENTRYPOINT ["terraform"]
