# PR Tag Builder

## Overview

PR Tag Builder is a tool that finds a pull request number for a commit.

## Prerequisites

Tool is called on postsubmit Prow jobs. You need Prow instance to use it.

## Installation

The tool is a part of the `test-infra` **prow-tools** image. It is copied from the **prow-tools** image and installed in almost all images under the `test-infra/prow/images/prow-tools` path.

## Usage

Tool doesn't accept any flags and arguments. All data required for the tool to work, is retrived from environment variable **`JOB_SPEC`**. This environment variable is set by Prow for postsubmit jobs.`

The tool fails on any error that prevents it from finding a valid PR number for a commit.

Tool will not work on presubmit jobs.

## Development

Changes in `prtagbuilder` source code will trigger Prow presubmit and postsubmit jobs. Jobs will run tests and build **prow-tools** image. Version of **prow-tools** image should be updated in test-infra images dockerfiles, by replacing image tag to match new version. This will trigger rebuild of images which copy from **prow-tools** image.
