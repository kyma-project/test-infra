# Buildpack Golang wit LibGit2 Docker Image

## Overview

This folder contains the LibGit2 library that is based on the Golang image.
Use it to build Golang components with `git2go` library.
This image is created for function-controller component which utilizes `libgit2`.

The image consists of:

- golang 1.16.6
- dep 0.5.4
- libgit2-dev 1.1.1

## Installation

To build the Docker image, run this command:

```bash
docker build -t buildpack-golang-libgit2 .
```
