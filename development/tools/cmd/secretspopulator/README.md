# Secrets Populator

## Overview

This command reads Secrets stored in a Gcloud bucket, decrypts them with a Key Management Service(KMS) key, and saves them as Kubernetes Secrets in a cluster.
You define Secrets to populate in the `secrets-def-file` input parameter.

## Usage

To run it, use:
```bash
env GOOGLE_APPLICATION_CREDENTIALS={path to service account file} go run main.go \
    -bucket={bucket_name} \
    -keyring={keyring} \
    -key={key} \
    -location={kms location} \
    -kubeconfig={path to kubeconfig} \
    -project={gcloud project name} \
    -secrets-def-file={path to file with definition of secrets to populate}
```

### Flags

See the list of available flags:

| Name                      | Required | Description                                                                                          |
| :------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **--bucket**              |   Yes    | The name of the Gcloud bucket name where Secrets are stored                                
| **--keyring**             |   Yes    | KMS key ring            
| **--key**                 |   Yes    | KMS key
| **--location**            |   Yes    | KMS location            
| **--kubeconfig**          |   Yes    | The path to the `kubeconfig` file that points to the Prow cluster    
| **--secrets-def-file**    |   Yes    | The path to the YAML file that defines Secrets to populate. See the `RequiredSecretsData` type to learn about the syntax of the file.   
| **--project**             |   Yes    | Gcloud project name   

### Environment variables

See the list of available environment variables:

| Name                                  | Required | Description                                                                                          |
| :------------------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **GOOGLE_APPLICATION_CREDENTIALS**    |    Yes   | The path to the application credentials. It requires KMS and storage roles.                            
