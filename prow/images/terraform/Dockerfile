FROM hashicorp/terraform:1.3.7

RUN apk add --no-cache unzip curl && \
    mkdir /tfplugins && \
    curl -fLSs -o /terraform-provider-google_4.50.0.zip "https://releases.hashicorp.com/terraform-provider-google/4.50.0/terraform-provider-google_4.50.0_linux_amd64.zip" && \
    unzip -q /terraform-provider-google_4.50.0.zip -d /tfplugins//terraform-provider-google_4.50.0 && \
    curl -fLSs -o tfcmt_linux_amd64.tgz "https://github.com/suzuki-shunsuke/tfcmt/releases/download/v4.0.1/tfcmt_linux_amd64.tar.gz" && \
    tar -xgf tfcmt_linux_amd64.tgz -C /usr/local/bin


WORKDIR $GOPATH
ENTRYPOINT ["tfcmt", "-owner", "$REPO_OWNER", "-repo", "$REPO_NAME", "-pr", "$PULL_NUMBER"]
