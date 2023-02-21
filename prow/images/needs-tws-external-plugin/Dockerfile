# This dockerfile is meant to use on externally built binary using command:
# GOOS=linux GOARCH=amd64 go build -o ./needs-tws .
# then use docker build -t needs-tws .
# NOT MEANT FOR AUTOBUILDS

FROM alpine:3.17.2

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit
COPY ../../../development/external-plugins/needs-tws/needs-tws /
RUN apk add --no-cache ca-certificates git && \
	chmod a+x /needs-tws
ENTRYPOINT ["/needs-tws"]
