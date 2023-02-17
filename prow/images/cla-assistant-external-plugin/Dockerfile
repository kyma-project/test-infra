# This dockerfile is meant to use on externally built binary using command
# in development/tools/external-plugins/cla-assistant:
# GOOS=linux GOARCH=amd64 go build -o ./plugin .
# Move the compiled binary file to this directory and execute:
# docker build -t cla-assistant .
# NOT MEANT FOR AUTOBUILDS

FROM alpine:3.17.2

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit
COPY ./plugin /
RUN apk add --no-cache ca-certificates git && \
	chmod a+x /plugin
ENTRYPOINT ["/plugin"]
