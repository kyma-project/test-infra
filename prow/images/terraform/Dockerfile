FROM hashicorp/terraform:1.3.7

RUN apk add --no-cache unzip curl && \
    mkdir /tfplugins && \
    mkdir /tfcmt && \
    mkdir /ghcomment && \
#    curl -fLSs -o /terraform-provider-google_4.50.0.zip "https://releases.hashicorp.com/terraform-provider-google/4.50.0/terraform-provider-google_4.50.0_linux_amd64.zip" && \
#    unzip -q /terraform-provider-google_4.50.0.zip -d /tfplugins//terraform-provider-google_4.50.0 && \
    curl -fLSs -o tfcmt_linux_amd64.tgz "https://github.com/suzuki-shunsuke/tfcmt/releases/download/v4.0.1/tfcmt_linux_amd64.tar.gz" && \
    tar -xzf tfcmt_linux_amd64.tgz -C /tfcmt && \
    curl -fLSs -o github-comment_5.0.2_linux_amd64.tgz "https://github.com/suzuki-shunsuke/github-comment/releases/download/v5.0.2/github-comment_5.0.2_linux_amd64.tar.gz" && \
    tar -xzf github-comment_5.0.2_linux_amd64.tgz -C /ghcomment

RUN PATH="/tfcmt:$PATH"


WORKDIR $GOPATH
