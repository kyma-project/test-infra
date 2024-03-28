# External Secrets Checker

## Overview

This command checks external Secrets synchronization status, and if every Secret has a corresponding external Secret.

## Usage

To run it, use:

```bash
go run main.go --kubeconfig=~/.kube/config --ignored-secrets "namespace/secretName,namespace/secretName2"
```

### Flags

See the list of available flags:

| Name                      | Required | Description                                                                                                           |
| :------------------------ | :------: | :-------------------------------------------------------------------------------------------------------------------- |
| **--kubeconfig**          |   Yes    | The path to the `kubeconfig` file needed to connect to a cluster.                                                     |
| **--ignored-secrets**     |    No    | The list of ignored Secrets. The Secrets are in the `namespace/secretName` format and are separated with commas.      |
| **--namespaces**          |    No    | The list of analyzed namespaces. The namespaces names are separated with commas. The program scans all namespaces if the namespaces list is empty. |
