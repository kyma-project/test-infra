FROM hashicorp/terraform:1.3.7

RUN apk add --no-cache unzip curl && \
    mkdir /tfplugins && \
    curl -fLSs -o /terraform-provider-google_4.50.0.zip "https://releases.hashicorp.com/terraform-provider-google/4.50.0/terraform-provider-google_4.50.0_linux_amd64.zip" && \
    unzip -q /terraform-provider-google_4.50.0.zip -d /tfplugins//terraform-provider-google_4.50.0


WORKDIR $GOPATH
ENTRYPOINT ["terraform"]
