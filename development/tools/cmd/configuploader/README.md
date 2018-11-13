# Config Uploader

## Overview

This command is responsible for uploading Prow configuration.

## Usage

To run, use:

```bash
go run cmd/configuploader/main.go --kubeconfig $HOME/.kube/config --plugin-config-path {pathToPluginsYaml}
```

## Parameters

| Name                 | Required | Description                                                                                         |
| :------------------- | :------: | :-------------------------------------------------------------------------------------------------- |
| --kubeconfig         |   Yes    | The path to the kubeconfig file, needed for connecting to the cluster.                              |
| --config-path        |    No    | The path to the `config.yaml` file. If set, the configuration will be uploaded.                     |
| --jobs-config-path   |    No    | The path to the directory with job configurations. If set, the job configurations will be uploaded. |
| --plugin-config-path |    No    | The path to the `plugins.yaml` file. If set, the plugins configuration will be uploaded.            |
