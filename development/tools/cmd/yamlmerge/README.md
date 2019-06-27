# YAML merge

## Overview

This command-line tool enables merging YAML files on a given path into one file. It simply creates a`starter.yaml` file from the component configurations under the `/prow/cluster/components` path, and generates a file that includes the content of `/prow/cluster/starter.yaml`.

## Usage

For security reasons, the dry-run mode is the default one.
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
| **--dryRun**              |    No    | The boolean value that controls the dry-run mode. Its default value is `true`.


