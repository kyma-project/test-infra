# YAML merge

## Overview

This commandline tool enables merging of YAML files on a given path into one file. It should be used to create a starter.yaml file from the components in the `/prow/cluster/components` path, to generate a file that is of the contents of `/prow/cluster/starter.yaml`.

## Usage

For safety reasons, the dry-run mode is the default one.
To run it, use:
```bash
go run main.go -path /path/to/components/folder -target /path/to/target/file
```

To turn the dry-run mode off, use:
```bash
go run main.go -path /path/to/components/folder -target /path/to/target/file -dryRun=false
```

### Flags

See the list of available flags:

| Name                      | Required | Description                                                                                          |
| :------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **--path**                |   Yes    | Path to the folder containing the yaml files to merge.
| **--target**              |   Yes    | Path of the file, that is target to collected files' merged content.
| **--dryRun**              |    No    | The boolean value that controls the dry-run mode. It defaults to `true`.


