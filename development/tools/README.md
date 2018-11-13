# Tools

## Overview

This project contains Go applications for test-infra repository.

## Prerequisites

Use the following tools to set up the project:

- Go
- Dep

## Usage

1.  Download all dependencies:

    ```bash
    dep ensure -vendor-only -v
    ```

2.  Run command from `cmd` directory, for example:

    ```bash
    go run cmd/configuploader/main.go --kubeconfig $HOME/.kube/config --plugin-config-path {pathToPluginsYaml}
    ```

    For details how to run commands, go to `cmd` directory.
