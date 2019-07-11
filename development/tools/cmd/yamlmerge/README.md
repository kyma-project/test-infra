# YAML merge

## Overview

This command line tool enables merging yaml files into one single file. For the operation to work, the yaml files must follow the same source path. 
For example, the following command allows the tool to merge the component configurations residing under the `/prow/cluster/components` path and place the resulting target file under `/prow/cluster/starter.yaml`. 
```bash
go run main.go -path /prow/cluster/components -target /prow/cluster/starter.yaml
```

>**NOTE**:  If the `starter.yaml` file does not already exist under a given path, it will be created.

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
| **--target**              |   Yes    | Path of the file which includes the content of the merged files.
| **--dryRun**              |    No    | The boolean value that controls the dry-run mode. Its default value is `true`.


