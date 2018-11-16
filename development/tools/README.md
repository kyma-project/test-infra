# Tools

## Overview

This project contains Go applications for the `test-infra` repository.

## Prerequisites

Use the following tools to set up the project:

- Go
- Dep

## Usage

1.  Download all dependencies:

    ```bash
    dep ensure -vendor-only -v
    ```

2.  Run a command from the `cmd` directory. See an example:

    ```bash
    go run cmd/configuploader/main.go --kubeconfig $HOME/.kube/config --plugin-config-path {pathToPluginsYaml}
    ```

    For details on flags to use with the command, go to the `cmd` directory.
