## Overview

The folder contains files that are directly used by Prow pipeline scripts.

## Directory structure

```
├── kube-dns-stub-domains-patch.yaml                          # Enables the build.kyma-project.io stubdomain and provides Google root DNS servers IPs.
├── limitrange-patch.yaml                                     # Increases the kyma-system Namespace maximum memory request for containers.
├── prometheus-operator-stackdriver-patch.yaml                # Injects the Stackdriver collector sidecar, sets metric filters, and enables scraping Stackdriver target.
├── prometheus-operator-additional-scrape-config.yaml         # Additional scrape configuration for Prometheus operator.
└── debug-container.yaml                                      # Create pod and configmap for collecting memory related statistics for k8s node and containers.
```

## Prometheus operator additional scrape configuration

Prometheus operator expects to have additional scrape configuration provided as a Secret. This Secret is appended to the Prometheus scrape config file.
Additional scrape config allows you to add scrape targets outside automatic scrape targets discovery mechanisms.
It is administrator's responsibility to provide syntactically correct scrape configuration.

## Troubleshoot OOM events with debug-container.

Standard OOM events logged in os logs and k8s node events provide details about allocated memory, PID and process name which was killed to release memory. Killed process, due to OOM activity not always is the same which consume too much memory. Even if killed process is the one which consumed too much memory, it's hard to match its name with pods running on k8s node. For that purpose use a debug-container.yaml. It contains configmap with bash script and pod where this script is running. A pod is able to access host `/proc` and `/sys` filesystems and docker daemon socket. Once scheduled, it will collect host and running containers memory stats. Apart memory stats, for all containers it will collect container PID, container name and all PIDs belong to the same cgroup. Collected data should be copied to the artifacts directory on GCP bucket with prowjob logs.

Example output from oom-debug conatiner.

```
...
Thu Mar 11 12:25:56 UTC 2021
Host memory stats
memory.capacity_in_bytes: 16796868608
memory.usage_in_bytes: 6298832896
memory.total_inactive_file: 4770684928
memory.working_set: 1528147968
memory.available_in_bytes: 15268720640
memory.available_in_kb: 14910860
memory.available_in_mb: 14561
Containers
[11098] Container PID and name: 11098:/k8s_POD_node-problem-detector-hzhkg_kube-system_3b9dd357-e9db-4fbc-811c-dcf4df6a2418_3
[11098] Container memory.usage_in_bytes: 835584
[11098] Container memory.max_usage_in_bytes: 8548352
[11098] Container memory.limit_in_bytes: 9223372036854771712
[11098] Container memory.usage_percentage: 0
[11098] Container memory.max_usage_percentage: 0
[11098] cgroup processes: 11098
...
```
### How to use debug-container

First apply debug-container.yaml to start pod. This should immediately follow cluster creation command
```
# run oom debug pod
kubectl apply -f "${TEST_INFRA_SOURCES_DIR}/prow/scripts/resources/debug-container.yaml"
```

Next copy script output to *$ARTIFACTS* location. This should be run from cleanup function to call it for success and failed prowjobs executions.
```
# copy oom debug pod output to artifacts directory
    kubectl cp default/oom-debug:/var/oom_debug "${ARTIFACTS}/oom_debug.txt"
```
