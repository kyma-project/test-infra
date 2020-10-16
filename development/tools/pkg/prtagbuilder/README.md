# PR Tag Builder

## Overview

PR Tag Builder is a tool that finds a pull request number for a commit.

## Prerequisites

The tool is called on postsubmit Prow jobs. You need a Prow instance to use it.

## Installation

The tool is a part of the `test-infra` **prow-tools** image. It is copied from the **prow-tools** image and installed in almost all images under the `test-infra/prow/images/prow-tools` path.

## Usage

The tool doesn't accept any flags and arguments. All data required for the tool to work is retrieved from the  **JOB_SPEC** environment variable. This environment variable is set by Prow for postsubmit jobs.`

The tool fails on any error that prevents it from finding a valid PR number for a commit.

The tool doesn't work on presubmit jobs.

## Development

Changes in the `prtagbuilder` source code trigger Prow pre-submit and postsubmit jobs. Jobs run tests and build the **prow-tools** image. The version of the **prow-tools** image should be updated in `test-infra` Dockerfile images by replacing the image tag to match the new version. This rebuilds the images copied from the **prow-tools** image.
