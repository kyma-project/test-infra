# Image Syncer

## Overview

TODO

## Usage

To run it, use:
```bash
go run main.go \ 
    --images-file={path to a yaml file containing sync definition} \
    --target-key-file={path to a json key file} \
    --dry-run=true
```

### Flags

```
Usage:
  image-syncer [flags]

Flags:
  -d, --dry-run                  dry run mode [SYNCER_DRY_RUN] (default true)
  -h, --help                     help for githubstats
  -i, --images-file string       yaml file containing list of images [SYNCER_IMAGES_FILE]
  -t, --target-key-file string   JSON key file used for authorization to target repo [SYNCER_TARGET_KEY_FILE]

exit status 1
```


### Environment variables

All flags can also be set using the environment variables:

| Name                           | Required | Description                                                           |
| :----------------------------- | :------: | :-------------------------------------------------------------------- |
| **SYNCER_IMAGES_FILE**         |    Yes   | The string value with a path to yaml file with sync definition.       |
| **SYNCER_TARGET_KEY_FILE**     |    Yes   | The string value with a path to json key file.                        |
| **SYNCER_DRY_RUN**             |    No    | The boolean value controlling the `dry run` mode.                     |
