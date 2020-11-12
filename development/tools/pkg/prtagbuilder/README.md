# PR Tag Builder

## Overview

PR Tag Builder is a tool that finds a pull request number for a commit.

## Prerequisites

The tool is a go binary. There are no prerequisites.

## Installation

The tool is a part of the `test-infra` **prow-tools** image. It is copied from the **prow-tools** image and installed in almost all images under the `test-infra/prow/images/prow-tools` path.

## Usage

All data required for the tool to work can be retrieved from the **JOB_SPEC** environment variable. This environment variable is set by Prow for all prowjobs. In this mode a tool will find pull request number for base SHA of branch for which prowjob is running.

Optionally prtagbuilder can be run with flags which instruct it to find pull request number for head of provided branch.

`prtagbuilder` accept following flags.

| Parameter name | Short name | Required | Description |
|----------------|------------|----------|-------------|
| **org** | o | No | Github owner name of repository to find PR number for. If provided, **repo** and **baseref** flags must be provided. |
| **repo** | r | No | Github repository to find PR number for. If provided, **org** and **baseref** flags must be provided. |
| **baseref** | b | No | Branch name to find a PR number for. If provided, **org** and **repo** flags must be provided. |
| **numberOnly** | O | No | Print only a PR number. By default print pr tag in format `PR-<pr number>` |

The tool fails on any error that prevents it from finding a valid PR number for a commit.

## Development

Changes in the `prtagbuilder` source code trigger Prow presubmit and postsubmit jobs. Jobs run tests and build the **prow-tools** image. The version of the **prow-tools** image should be updated in `test-infra` Dockerfile images by replacing the image tag to match the new version. This rebuilds the images copied from the **prow-tools** image.
