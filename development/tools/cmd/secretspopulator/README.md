# Secrets Populator

## Overview

This command reads secrets stored in Gcloud bucket, decrypt it with KMS key, and stores them as a k8s Secrets in a cluster.
By default, k8s secret data key is set to "service-account.json". It can be overriden, by specifying object metadata with key `override-secret-data-key`.

## Usage

To run it, use:

```bash
env GOOGLE_APPLICATION_CREDENTIALS={path to service account file} go run main.go \ 
    -bucket={bucket_name} 
    -keyring={keyring} 
    -key={key} 
    -location={kms location} 
    -kubeconfig={path to kubeconfig}
    -project={gcloud project name}
    -secrets-def-file={path to file with definition of secrets to populate}
```


### Flags

See the list of available flags:

| Name                      | Required | Description                                                                                          |
| :------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **--bucket**              |   Yes    | Gcloud bucket name where secrets are stored                                
| **--keyring**             |   Yes    | KMS keyring            
| **--key**                 |   Yes    |  KMS key
| **--location**            |   Yes    |  KMS location            
| **--kubeconfig**          |   Yes    | Path to kubeconfig file that points to Prow cluster    
| **--secrets-def-file**    |   Yes    | Path to yaml file defining secrets to populate. See `RequiredSecretsData` type to learn format of the file.   
| **--project**             |   Yes    | Gcloud project name   

### Environment variables

See the list of available environment variables:

| Name                                  | Required | Description                                                                                          |
| :------------------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **GOOGLE_APPLICATION_CREDENTIALS**    |    Yes   | Path to application credentials.                              
