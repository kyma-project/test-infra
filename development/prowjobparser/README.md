# prowjobparser

## Overview

The prowjobparser is a helper tool which parses all ProwJobs under the provided path, matches them against the provided label filters, and prints matching ProwJob names to the standard output.

## Usage

### CLI parameters

The prowjobsparser accepts the following command line parameters:

|Parameter | Shorthand | Description |
|-----------|-----------|------------|
| **configpath** | **-c** | Path to prow config yaml file. |
| **jobpath** | **-j** | Path to directory containing yaml files with prowjobs. |
| **includepreset** | **-i** | Preset name which must be added to prowjob. Accept multiple param instances. | 
| **excludepreset** | **-e** | Preset name which shouldn't be added to prowjob. Accept multiple param instances. | 
