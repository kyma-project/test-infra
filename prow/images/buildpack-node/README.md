# Buildpack Node.js Docker Image

## Overview

This folder contains the Buildpack Node.js image that is based on the Bootstrap image. Use it to build Node.js components.

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
