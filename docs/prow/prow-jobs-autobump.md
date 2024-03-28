
# Image Autobump 

This document provides an overview of autobump Prow Jobs. 

## Overview

[Generic-autobumper](https://github.com/kubernetes/test-infra/tree/master/prow/cmd/generic-autobumper) tool allows for automatic upgrading of Docker images URLs in the `values.yaml` files to their latest versions; the images have to be specified as Docker images URLs.

# Autobumper Job Template

The following template data file can be used to generate the autobumper job for Kyma repository:

```yaml
  - from: templates/generic.tmpl
    render:
      - to: ../prow/jobs/kyma/kyma-autobump.yaml
        localSets:
          github_token_mounts:
            labels:
              preset-autobump-bot-github-token: "true"
        jobConfigs:
          - repoName: kyma-project/kyma
            jobs:
              - jobConfig:
                  name: kyma-autobump
                  cron: "30 * * * 1-5"
                  image: gcr.io/k8s-prow/generic-autobumper:v20220524-dfb23cb2d1
                  command: generic-autobumper
                  args:
                    - --config=prow/autobump-config/kyma-autobump-config.yaml
                inheritedConfigs:
                  local:
                    - github_token_mounts
                  global:
                    - jobConfig_default
                    - jobConfig_postsubmit
                    - pubsub_labels
```
