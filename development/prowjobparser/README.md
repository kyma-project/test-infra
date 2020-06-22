# prowjobparser

## Overview

Prowjobparser is a helper tool which parse all prowjobs under provided path, match them against provided labels filters and print matching prowjob names to stdout.

## Usage

### CLI parameters

The prowjobsparser accepts the following command line parameters:

|Parameter | Shorthand | Description |
|-----------|-----------|------------|
| **configpath** | **-c** | Path to prow config yaml file. |
| **jobpath** | **-j** | Path to directory containing yaml files with prowjobs. |
| **includepreset** | **-i** | Preset name which must be added to prowjob. Accept multiple param instances. | 
| **excludepreset** | **-e** | Preset name which shouldn't be added to prowjob. Accept multiple param instances. | 
