# Due to scoping of the ARG, the ARG used in the FROM statement
# must be defined on top of the Dockerfile, before first FROM statement.
# See: https://github.com/rcjsuen/dockerfile-language-server/issues/163
# See: https://github.com/rcjsuen/dockerfile-language-server/issues/153
# See: https://docs.docker.com/reference/dockerfile/#understand-how-arg-and-from-interact
ARG BASE_ALPINE_IMAGE
FROM google/cloud-sdk:524.0.0 AS base

ARG AWS_IAM_AUTHENTICATOR_VERSION="0.6.11"
RUN curl -fsSL \
    https://github.com/kubernetes-sigs/aws-iam-authenticator/releases/download/v${AWS_IAM_AUTHENTICATOR_VERSION}/aws-iam-authenticator_${AWS_IAM_AUTHENTICATOR_VERSION}_linux_amd64 -o /aws-iam-authenticator && \
    chmod +x /aws-iam-authenticator


FROM ${BASE_ALPINE_IMAGE}

COPY --from=base /usr/bin/gke-gcloud-auth-plugin /usr/local/bin/gke-gcloud-auth-plugin
COPY --from=base /aws-iam-authenticator /usr/local/bin/aws-iam-authenticator
