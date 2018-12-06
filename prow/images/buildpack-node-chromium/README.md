# Node.js buildpack with Chromium Docker image

## Overview

This folder contains the Buildpack with Chromium installed that is based on the [buildpack-node](../buildpack-node) image. Use it to run tests requiring Chromium.

The image consists of:

- nodejs
- eslint
- tslint
- prettier
- whitesource
- chromium

## Installation

To build the Docker image, run this command:

```bash
docker build .
```
