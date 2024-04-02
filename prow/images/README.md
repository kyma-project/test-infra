# Images

> DEPRECATED: Use the [`images`](/images) directory instead.

## Overview

This folder contains a list of images used in Prow jobs.

### Project Structure

<!-- Update the folder structure each time you modify it. -->

The structure of the folder looks as follows:

```
  ├── bootstrap             # The generic image that contains Docker and gcloud            
  ├── bootstrap-helm        # The image that contains gcloud, Docker, and Helm
  ├── buildpack-golang      # The image for building Golang components
  ├── buildpack-node        # The image for building Node.js components
  ├── buildpack-java        # The image for building Java components
  ├── cleaner               # The image with a script for cleaning SSH keys on service accounts in Google Cloud Storage
  ├── golangci-lint         # The image contains a golangci-lint binary with wrapper script for scanning all folders containing go.mod files
  └── whitesource-scanner   # The image for performing whitesource scans
```
