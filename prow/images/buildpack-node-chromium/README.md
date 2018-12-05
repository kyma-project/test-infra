# Buildpack Node.js + Chromium Docker Image

## Overview

This folder contains the Buildpack with Node.js and Chromium installed that is based on the Bootstrap image. Use it to run tests requiring Puppeteer API.

The image consists of:

- nodejs
- eslint
- tslint
- prettier
- whitesource
- chromium

## Installation

To build the Docker image, run this command:

```bash
docker build .
```
