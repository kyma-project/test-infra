# gardener-rotate

## Overview

gardener-rotate tool allows to generate a new access token for Gardener service accounts and update kubeconfig stored in the Secret Manager.

## Usage

To run gardener-rotate, use:
```bash
go run main.go \ 
    --service-account={PATH_TO_A_JSON_KEY_FILE} \
    --config-file={PATH_TO_A_YAML_FILE_CONTAINING_CONFIGURATION} \
    --kubeconfig={PATH_TO_A_KUBECONFIG_FILE} \
    --dry-run=true
```


### Configuration file

gardener-rotate takes as an input parameter a file having the following structure: 

```yaml
serviceAccounts:
  - serviceAccount: "sa-neighbor-robot" # Kubernetes service account name
    namespace: "garden-neighbors" # Kubernetes service account namespace
    duration: 5184000 # vailidity of the new token in seconds
    secret: "trusted_default_gardener-neighbors-kubeconfig" # name of the GCP secret
    project: "sap-kyma-prow" # name of the GCP project with the Secret Manager
    keepOld: false # should old versions of the secret be disabled
```


# Gardener-rotate flags

See the list of flags available for the `promote` command:

| Name                      | Required | Description                                                                                          |
| :------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |

| **--service-account**     |   Yes    | Path to GCP service account credentials file.|
| **--kubeconfig**          |   Yes    | Path to the Gardener kubeconfig file.|
| **--config-file**         |   Yes    | Path to the `gardener-rotate` configuratino file.|
| **--dry-run**             |   No     | The boolean value that controls the dry-run mode. It defaults to `true`.|
| **--debug**               |   No     | The boolean value that controls the debug output.|
