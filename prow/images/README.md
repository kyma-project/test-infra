# Images

## Overview

This folder contains a list of images used in Prow jobs.

### Project structure

<!-- Update the folder structure each time you modify it. -->

The structure of the folder looks as follows:

```
  ├── bootstrap             # A generic image that contains Docker and gcloud.            
  ├── buildpack-golang      # An image for building Golang components.
  ├── buildpack-node        # An image for building Node.js components.
  └── cleaner               # An image with a script for cleaning SSH keys on service accounts in Google Cloud Storage.  
```
