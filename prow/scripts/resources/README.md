## Overview

The folder Contains files which are directly used by prow pipeline scripts.

## Content

```
├── kube-dns-stub-domains-patch.yaml                          # Enable stubdomain build.kyma-project.io and provide Google root DNS servers IPs.
├── limitrange-patch.yaml                                     # Increase kyma-system namespace max memory request for containers.
├── prometheus-operator-stackdriver-patch.yaml                # Inject stackdriver collector sidecar, set metrics filters, enable scraping stackdriver target.
└── prometheus-operator-additional-scrape-config.yaml         # Additionall scrape config for prometheus oeprator.
```

## Prometheus operator additional scrape config

Prometheus operator expect to have additional scrape config provided as one secret. This secret is appended to the Prometheus scrape config file as it is.
Additional scrape config allow add scrape targets outside automatic scrape targets discovery mechanisms.
It's admin responsibility to provide syntactically correct scrape configuration.
