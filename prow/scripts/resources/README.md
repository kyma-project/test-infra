## Overview

The folder contains files that are directly used by Prow pipeline scripts.

## Directory structure

```
├── kube-dns-stub-domains-patch.yaml                          # Enables the build.kyma-project.io stubdomain and provides Google root DNS servers IPs.
├── limitrange-patch.yaml                                     # Increases the kyma-system Namespace maximum memory request for containers.
├── prometheus-operator-stackdriver-patch.yaml                # Injects the Stackdriver collector sidecar, sets metric filters, and enables scraping Stackdriver target.
├── prometheus-operator-additional-scrape-config.yaml         # Additional scrape configuration for Prometheus operator.
└── debug-container.yaml                                      # Objects needed to run debug commando pod. 
```

## Prometheus operator additional scrape configuration

Prometheus operator expects to have additional scrape configuration provided as a Secret. This Secret is appended to the Prometheus scrape config file.
Additional scrape config allows you to add scrape targets outside automatic scrape targets discovery mechanisms.
It is administrator's responsibility to provide syntactically correct scrape configuration.

## Debug commando

Debug commando is a pod with debugging tools. It contains an oomfinder container which listens for oom events published by the containerisation engine. Oomfinder requires a privileged context. To allow this, the debug commando pod is running under a `gardener.cloud:psp:privileged` policy. To use debug commando, follow these steps: 
* Make sure your Prow job calls [utils::debug_oom](https://github.com/kyma-project/test-infra/blob/732e1fc8cc887d4328ce457c7af9566fae79be97/prow/scripts/lib/utils.sh) just after creating a k8s cluster.
* Make sure your Prow job calls [utils::oom_get_output](https://github.com/kyma-project/test-infra/blob/732e1fc8cc887d4328ce457c7af9566fae79be97/prow/scripts/lib/utils.sh) as a last step before cluster cleanup begins.
* Label your Prow job with `preset-debug-commando-oom: "true"`.
