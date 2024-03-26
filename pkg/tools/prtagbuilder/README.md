# PR Tag Builder

## Overview

PR Tag Builder is a tool that finds a pull request (PR) number for a commit.

## Prerequisites

The tool is a Go binary. There are no prerequisites.

## Installation

The tool is a part of the `test-infra` **prow-tools** image. It is copied from the **prow-tools** image and installed in almost all images under the `test-infra/prow/images/prow-tools` path.

## Usage

You can retrieve all data required for the tool to work from the **JOB_SPEC** environment variable. This environment variable is set by Prow for all Prow jobs. In this mode, the tool finds a PR number for the base SHA of the branch for which the Prow job is running.

Optionally, PR Tag Builder can be run with flags that instruct it to find a PR number for the head of the provided branch.

PR Tag Builder accepts the following flags:

| Full name | Short name | Required | Description |
|----------------|------------|----------|-------------|
| **org** | o | No | GitHub owner name of the repository to find a PR number for. If provided, you must also specify the **repo** and **baseref** flags. |
| **repo** | r | No | GitHub repository to find a PR number for. If provided, you must also specify the **org** and **baseref** flags. |
| **baseref** | b | No | Branch name to find a PR number for. If provided, you must also specify the **org** and **repo** flags. |
| **numberonly** | O | No | Parameter that prints a PR number. By default, the tool prints a PR tag in the `PR-{PR_NUMBER} format.` |

The tool fails on any error that prevents it from finding a valid PR number for a commit.

## Development

Changes in the `prtagbuilder` source code trigger Prow presubmit and postsubmit jobs. Jobs run tests and build the **prow-tools** image. The version of the **prow-tools** image should be updated in `test-infra` Dockerfile images by replacing the image tag to match the new version. This rebuilds the images copied from the **prow-tools** image.
