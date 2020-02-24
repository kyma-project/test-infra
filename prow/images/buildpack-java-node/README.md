# Buildpack Node.js Docker Image

## Overview

This folder contains the Buildpack for Java & Node.js image that is based on the Bootstrap image. This image is used for whitesource scans.

The image consists of:

- buildpack-java
- additional instructions from buildpack-node

## Installation

To build the Docker image, run this command:

```bash
docker build buildpack-java-node .
```
