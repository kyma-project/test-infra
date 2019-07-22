Fluentd-gcp

 ## Overview

This directory contains the source files needed to make a Docker image
that collects Docker container log files using [Fluentd][fluentd]
and sends them to [Stackdriver Logging][stackdriverLogging].

 The image consists of:

 - fluent-gcp v19.6.1

 ## Installation

 To build the Docker image, run this command:

 ```bash
make build-image
```
# Collecting Docker Log Files with Fluentd and sending to GCP.

