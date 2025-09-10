# Image Detector

## Overview

Image Detector is a tool for updating the security scanner config with the list of images.

## Key Features:

Image Detector:
* Extracts image URLs from various file types.
* Updates the list of image URLs in the security scanner config file.

## Usage

```
Usage of image-detector:
  --terraform-dir string
    path to the directory containing Terraform files
  --sec-scanner-config
    path to the security scanner config field (Required)
  --autobump-config
    path to the config for autobumper for security scanner config
  --github-token-path
    path to github token for fetching inrepo config (default: "/etc/github/token")
```
