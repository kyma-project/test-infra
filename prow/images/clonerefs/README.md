# clonerefs

## Overview

This folder contains improved version of clonerefs that is based on the original [clonerefs](https://github.com/kubernetes/test-infra/tree/master/prow/cmd/clonerefs) image.
This version checks if the network (DNS) in a cluster is up before cloning the repository.

## Installation

To build the Docker image, run this command:

```bash
make build-image
```
