# Render Templates

## Overview

Render Template is a tool that reads the configuration in the [`config.yaml`](../../../../templates/config.yaml) file to generate output files, such as Prow component jobs.

The `config.yaml` file specifies the following for the Render Template:
- Templates it must use to generate the files
- The name and location of the output files
- Values it must use to generate the files

## Usage

To run this tool, use this command:

```bash
go run development/tools/cmd/rendertemplates/main.go --config templates/config.yaml
```

### Flags

This tool uses one flag:

| Name | Required | Description                                                                                          |
| ------------------------ | :------: | --------------------------------------------------------------------------------------------------- |
| **--config**  |   Yes    | Specifies the location of configuration file used to generate output files by the Render Templates tool. |        
