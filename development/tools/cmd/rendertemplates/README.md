# Render Templates

## Overview

The Render Templates is a tool that reads the configuration in the [`templates/config.yaml`](../../../../templates/config.yaml) file to generate output files, such as Prow component jobs.

The `config.yaml` file specifies the following for the Render Templates:
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
| **&#x2011;&#x2011;config**  |   Yes    | Generates output files based on the definition of the configuration file. |        
