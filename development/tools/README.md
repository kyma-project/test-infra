# Tools

## Overview

This project contains Go applications for the `test-infra` repository.

## Prerequisites

Use the following tools to set up the project:

- Go

## Usage

1.  Run a command from the `cmd` directory. See an example:

    ```bash
    go run cmd/configuploader/main.go --kubeconfig $HOME/.kube/config --plugin-config-path {pathToPluginsYaml}
    ```

    For details on flags to use with the command, go to the `cmd` directory.
