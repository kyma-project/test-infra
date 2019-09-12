# Buildpack Golang Toolbox Docker Image

## Overview

This folder contains the Buildpack Golang Toolbox image that is based on the Buildpack Golang image. Use it to build Golang components using the versioned set of build tools.

In addition to the version introduced in Buildpack Golang, this image adds these tools:
- goimports
- errcheck
- golint
- mockery
- failery

They are downloaded using `go get` so their version will change to the newest one every time image is rebuilt. 

## Installation

To build the Docker image, run this command:

```bash
docker build -t buildpack-golang-toolbox .
```
