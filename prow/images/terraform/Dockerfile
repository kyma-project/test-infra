FROM hashicorp/terraform:1.5.7

RUN apk add --no-cache unzip curl && \
    mkdir /tfplugins && \
    mkdir /tfcmt && \
    mkdir /ghcomment && \
    #    curl -fLSs -o /terraform-provider-google_4.50.0.zip "https://releases.hashicorp.com/terraform-provider-google/4.50.0/terraform-provider-google_4.50.0_linux_amd64.zip" && \
    #    unzip -q /terraform-provider-google_4.50.0.zip -d /tfplugins//terraform-provider-google_4.50.0 && \
    curl -fLSs -o tfcmt_linux_amd64.tgz "https://github.com/suzuki-shunsuke/tfcmt/releases/download/v4.7.0/tfcmt_linux_amd64.tar.gz" && \
    tar -xzf tfcmt_linux_amd64.tgz -C /tfcmt && \
    rm tfcmt_linux_amd64.tgz && \
    curl -fLSs -o github-comment_6.0.1_linux_amd64.tgz "https://github.com/suzuki-shunsuke/github-comment/releases/download/v6.0.1/github-comment_6.0.1_linux_amd64.tar.gz" && \
    tar -xzf github-comment_6.0.1_linux_amd64.tgz -C /ghcomment && \
    rm github-comment_6.0.1_linux_amd64.tgz

RUN PATH="/tfcmt:$PATH"


WORKDIR $GOPATH
