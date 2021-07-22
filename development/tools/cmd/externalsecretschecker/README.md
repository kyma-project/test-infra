# External secrets checker

## Overview

This command checks external secrets synchronization status, and if every secret have corresponding external secret.

## Usage

To run it, use:

```bash
go run main.go --kubeconfig=~/.kube/config --ignored-secrets "namespace/secretName,namespace/secretName2"
```

### Flags

See the list of available flags:

| Name                      | Required | Description                                                                                             |
| :------------------------ | :------: | :------------------------------------------------------------------------------------------------------ |
| **--kubeconfig**          |   Yes    | The path to the `kubeconfig` file, needed to connect to a cluster.                                      |
| **--ignored-secrets**     |    No    | List of ignored secrets. List contains of secrets in `namespace/secretName` format, separated by comma. |
| **--namespaces**          |    No    | List of analyzed namespaces. The program scans all namespaces if the list is empty.                     |
