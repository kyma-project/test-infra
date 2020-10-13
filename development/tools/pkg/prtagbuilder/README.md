# PR-tag builder

## Overview

**`prtagbuilder`** is a tool which will find a pull request number for a commit.

## Prerequisites

Tool is called on postsubmit Prow jobs. You need Prow instance to use it.

## Installation

Tool is a part of test-infra **prow-tools** image. It is installed in almost all test-infra images under `/prow-tools` path by copying from **prow-tools** image.

## Usage

Tool doesn't accept any flags and arguments. All data required for the tool to work, is retrived from environment variable **`JOB_SPEC`**. This environment variable is set by Prow for postsubmit jobs.`

Tool will fail on any error which will prevent it from finding valid PR number for a commit.

Tool will not work on presubmit jobs.

## Development

Changes in `prtagbuilder` source code will trigger Prow presubmit and postsubmit jobs. Jobs will run tests and build **prow-tools** image. Version of **prow-tools** image should be updated in test-infra images dockerfiles, by replacing image tag to match new version. This will trigger rebuild of images which copy from **prow-tools** image.

