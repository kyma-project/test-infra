# image-detector

Image Detector is a tool used to update security scanner config with list of images that lives in our prow cluster. To achive that it receives paths to files used to deploy prow or it's components.

### Key Features:

* Extract image urls from various file types
* Update list of image urls in security scanners config file

## Usage

```
Usage of image-detector:
  --prow-config string
    path to the prow config file (Required)
  --prow-jobs-dir string
    path to the directory which contains prow job files (Required)
  --terraform-dir string
    path to the directory containing terraform files (Required)
  --sec-scanner-config
    path to the security scanner config field (Required)
  --kubernetes-dir string
    path to the directory containing kubernetes deployments (Required)
  --tekton-catalog string
    path to the tekton catalog directory (Required)
```

