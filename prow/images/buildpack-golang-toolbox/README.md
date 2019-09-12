# Buildpack Golang Toolbox Docker Image

## Overview

This folder contains the Buildpack Golang Toolbox image that is based on the Buildpack Golang image. Use it to build Golang components using the versioned set of build tools.

In addition to the version introduced in Buildpack Golang, this image adds these tools:
- goimports
- errcheck
- golint
- mockery
- failery

> **CAUTION:** The `go get` command that downloads these tools always fetches the latest packages. Because of that, the image may contain different versions of the tools every time it is rebuilt.

## Installation

To build the Docker image, run this command:

```bash
docker build -t buildpack-golang-toolbox .
```
