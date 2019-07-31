# Vulnerability Scanner

## Overview

This folder contains the WhiteSource Unified Agent image that is based on the Java Buildpack image. Use it to perform WhiteSource vulnerability scans.
Scans are scheduled on every Monday at 4am UTC

The image contains `whitesource agent v19.6.1`.

## Installation

To build the Docker image, run this command:

```bash
make build-image
```
