# Buildpack Node Docker Image

## Overview

This folder contains the Buildpack Node image that is based on the Bootstrap image. Use it to build node components.

The image consists of:

- nodejs
- eslint
- tslint
- prettier
- whitesource

## Installation

To build the Docker image, run this command:

```bash
docker build buildpack-node .
```
