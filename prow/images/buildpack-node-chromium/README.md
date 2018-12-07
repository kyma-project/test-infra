# Node.js buildpack with Chromium Docker image

## Overview

This folder contains the Node.js buildpack with Chromium installed. It is based on the [buildpack-node](../buildpack-node) image, and you can use it to run tests requiring Chromium.

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
