# prowjobparser

## Overview

The prowjobparser is a helper tool which parses all ProwJobs under the provided path, matches them against the provided label filters, and prints matching ProwJob names to the standard output.

## Usage

### CLI parameters

The prowjobsparser accepts the following command line parameters:

|Parameter | Shorthand | Description |
|-----------|-----------|------------|
| **configpath** | **-c** | Path to the Prow config YAML file. |
| **jobpath** | **-j** | Path to the directory containing YAML files with ProwJobs. |
| **includepreset** | **-i** | Preset name which must be added to ProwJobs. Accepts multiple-parameter instances. | 
| **excludepreset** | **-e** | Preset name which must not be added to ProwJobs. Accepts multiple-parameter instances. | 
