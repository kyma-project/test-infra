## Overview

The folder contains files that are directly used by Prow pipeline scripts.

## Directory structure

```
├── kube-dns-stub-domains-patch.yaml                          # Enables the build.kyma-project.io stubdomain and provides Google root DNS servers IPs.
├── limitrange-patch.yaml                                     # Increases the kyma-system Namespace maximum memory request for containers.
├── prometheus-operator-stackdriver-patch.yaml                # Injects the Stackdriver collector sidecar, sets metric filters, and enables scraping Stackdriver target.
└── prometheus-operator-additional-scrape-config.yaml         # Additional scrape configuration for Prometheus operator.
```

## Prometheus operator additional scrape configuration

Prometheus operator expects to have additional scrape configuration provided as a Secret. This Secret is appended to the Prometheus scrape config file.
Additional scrape config allows you to add scrape targets outside automatic scrape targets discovery mechanisms.
It is administrator's responsibility to provide syntactically correct scrape configuration.
