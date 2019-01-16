# Buildpack Golang Docker Image

## Overview

This folder contains the Buildpack Golang image that is based on the Bootstrap image. Use it to build Golang components.

The image consists of:

- golang 1.11.4
- dep 0.5.0

## Installation

To build the Docker image, run this command:

```bash
docker build -t buildpack-golang .
```
