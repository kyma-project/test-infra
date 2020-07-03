# prowjobparser

## Overview

The prowjobparser is a helper tool which parses all ProwJobs under the provided path, matches them against the provided label filters, and prints matching ProwJob names to the standard output.

## Usage

To use the prowjobparser, navigate to its directory and run it with appropriate parameters. See the example:

```bash
go run main.go \
  -c ../../prow/config.yaml \
  -j ../../prow/jobs \
  -i preset-sa-gke-kyma-integration \
  -e preset-sa-kyma-artifacts
```

### CLI parameters

The prowjobsparser accepts the following command line parameters:

|Parameter | Shorthand | Description |
|-----------|-----------|------------|
| **configpath** | **-c** | Path to the Prow config YAML file. |
| **jobpath** | **-j** | Path to the directory containing YAML files with ProwJobs. |
| **includepreset** | **-i** | Preset name which must be added to ProwJobs. Accepts multiple-parameter instances. | 
| **excludepreset** | **-e** | Preset name which must not be added to ProwJobs. Accepts multiple-parameter instances. | 
