# Node.js Docker Image Buildpack

## Overview

This folder contains the Buildpack for Java and Node.js image that is based on the Bootstrap image. This image is used for whitesource scans.

The image consists of:

- `buildpack-java`
- Additional instructions reused from the `buildpack-node` image

## Installation

To build the Docker image, run this command:

```bash
docker build buildpack-java-node
```
