## Overview

The folder contains files that are directly used by Prow pipeline scripts.

## Directory structure

```
├── kube-dns-stub-domains-patch.yaml                          # Enable stubdomain build.kyma-project.io and provide Google root DNS servers IPs.
├── limitrange-patch.yaml                                     # Increases the kyma-system Namespace maximum memory request for containers.
├── prometheus-operator-stackdriver-patch.yaml                # Injects the Stackdriver collector sidecar, sets metric filters, and enables scraping Stackdriver target.
└── prometheus-operator-additional-scrape-config.yaml         # Additionall scrape configuration for Prometheus operator.
├── kube-dns-stub-domains-patch.yaml                          # Enables stubdomain build.kyma-project.io and provides Google root DNS servers IPs.
├── limitrange-patch.yaml                                     # Increase kyma-system namespace max memory request for containers.
├── prometheus-operator-stackdriver-patch.yaml                # Inject stackdriver collector sidecar, set metrics filters, enable scraping stackdriver target.
└── prometheus-operator-additional-scrape-config.yaml         # Additionall scrape config for prometheus oeprator.
```

## Prometheus operator additional scrape configuration

Prometheus operator expects to have additional scrape configuration provided as a Secret. This Secret is appended to the Prometheus scrape config file.
Additional scrape config allows you to add scrape targets outside automatic scrape targets discovery mechanisms.
It is administrator's responsibility to provide syntactically correct scrape configuration.
