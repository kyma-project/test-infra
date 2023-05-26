# Image Detector

## Overview

Image Detector is a tool for updating the security scanner config with the list of images in the Prow cluster. To achieve that, it receives paths to files used to deploy Prow or its components.

## Key features:

Image Detector:
* Extracts image URLs from various file types
* Updates the list of image URLs in the security scanner config file

## Usage

```
Usage of image-detector:
  --prow-config string
    path to the Prow config file (Required)
  --prow-jobs-dir string
    path to the directory which contains Prow job files (Required)
  --terraform-dir string
    path to the directory containing Terraform files (Required)
  --sec-scanner-config
    path to the security scanner config field (Required)
  --kubernetes-dir string
    path to the directory containing Kubernetes deployments (Required)
  --tekton-catalog string
    path to the Tekton catalog directory (Required)
```

