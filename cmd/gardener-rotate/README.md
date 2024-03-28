# gardener-rotate

## Overview

The gardener-rotate tool allows you to generate a new access token for the Gardener service accounts and update kubeconfig stored in Secret Manager.

The rotation process consists of the following steps:
1. Connect to the Gardener cluster using the provided kubeconfig file.
2. For each service account defined in the config file:  
    i. Generate a new Gardener access token.  
    ii. Update the kubeconfig stored in the Secret Manager secret with the generated access token.  
    iii. Disable the old versions of the secret.

## Usage

To run gardener-rotate, use:
```bash
go run main.go \ 
    --service-account={PATH_TO_A_JSON_KEY_FILE} \
    --config-file={PATH_TO_A_YAML_FILE_CONTAINING_CONFIGURATION} \
    --kubeconfig={PATH_TO_A_KUBECONFIG_FILE} \
    --dry-run=true
```


### Configuration File

As an input parameter, gardener-rotate takes a file with the following structure: 

```yaml
serviceAccounts:
  - k8sServiceAccount: "sa-neighbor-robot" # Kubernetes service account name to rotate
    k8sNamespace: "garden-neighbors" # Kubernetes service account namespace
    k8sDuration: 5184000 # vailidity of the new token in seconds
    gcpSecretManagerSecretName: "trusted_default_gardener-neighbors-kubeconfig" # name of the Google Cloud secret where the kubeconfig is stored
    gcpProjectName: "sap-kyma-prow" # name of the Google Cloud project with Secret Manager
    gcpKeepOld: false # should old versions of the Google Cloud secret be disabled, false by default
```


# Gardener-Rotate Flags

See the list of flags available for the `promote` command:

| Name                      | Required | Description                                                                                          |
| :------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **--service-account**     |   Yes    | Path to the Google Cloud service account credentials file. This credential is used to access Secret Manager.|
| **--kubeconfig**          |   Yes    | Path to the Gardener kubeconfig file. This credential is used for token rotation.|
| **--config-file**         |   Yes    | Path to the `gardener-rotate` configuration file.|
| **--dry-run**             |   No     | The boolean value that controls the dry-run mode. It defaults to `true`.|
| **--cluster-name**        |   No     | Specifies the name of the cluster used in the generated kubeconfig.|
