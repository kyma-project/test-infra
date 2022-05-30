
# Image autobump 

This document provides an overview of autobump Prow jobs.  

## Overview

Generic-autobumper tool allows for automatic bump of Docker images in files to their latest versions. The images have to be specified as a Docker repository URL.

## Autobumper config file

Following template data file can be used to generate autobumper config for Kyma repository:

```yaml
templates:
  # generate autobump configuration
  - from : templates/autobump-config.yaml
    render:
      - to: ../prow/autobump-config/kyma-autobump-config.yaml
        values:
          org: kyma-project
          repo: kyma
          included_paths:
            - resources
          excluded_paths:
            - resources/rafter
          extra_files:
           - "non_yaml_file.go"
```

This template will update images in yaml files stored in `resources`, except `resources/rafter`. Additional non-YAML files have to be specified in `extra_files` list.

# Autobumper job template

Following template data file can be used to generate autobumper job for Kyma repository:

```yaml
  - from: templates/generic.tmpl
    render:
      - to: ../prow/jobs/kyma/kyma-autobump.yaml
        localSets:
          github_token_mounts:
            labels:
              preset-bot-github-token: "true"
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
                    - disable_testgrid
```
