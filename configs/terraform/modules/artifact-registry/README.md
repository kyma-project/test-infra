# gcp-artifact-registry-terraform

This is the GCP image registry creator tool. Use the registry to publish modules that should be accessible to internal SAP teams.

## Usage

Configure Artifact Registry related values in the `terraform.tfvars` file. 
> **CAUTION:** Do not delete or update the existing registry-related data set unless you have the knowledge necessary to perform the operations.


You can configure multiple artifact registries as a list of objects in `artifact_registry_collection`.

```terraform
artifact_registry_collection = {
  modules-internal={
    name                   = "modules-internal"
    owner                  = "neighbors"
    type                   = "production"
    reader_serviceaccounts = ["klm-controller-manager@sap-ti-dx-kyma-mps-dev.iam.gserviceaccount.com", "klm-controller-manager@sap-ti-dx-kyma-mps-stage.iam.gserviceaccount.com", "klm-controller-manager@sap-ti-dx-kyma-mps-prod.iam.gserviceaccount.com"]
    writer_serviceaccount  = "kyma-submission-pipeline@kyma-project.iam.gserviceaccount.com"
  },
}
```
If you want to create a new Artifact registry, copy an existing registry data and then modify the **required** parameters according to your needs. 
> **CAUTION:** Do not delete or update the existing registry-related data set unless you have the knowledge necessary to perform the operations.

```terraform
artifact_registry_collection = {
    ...
  <your registry's name>={
    name                   = "<your registry's name>"
    owner                  = "<registry owner>"
    type                   = "<type: development or production>"
    reader_serviceaccounts = ["<service account 1>", "<service account 2>"]
  },
  ...
}
```
Additionally, you can define optional parameters. Here are all the parameters you can use:
| Parameter              | Description                                                             | Type         | Required | Default value |
|------------------------|-------------------------------------------------------------------------|--------------|----------|---------------|
| **name**                   | Artifact Registry name                                                  | string       | x        |               |
| **owner**                  | Registry Owner Team                                                     | string       | x        |               |
| **type**                   | Environment type (development, production)                              | string       | x        |               |
| **reader_serviceaccounts** | List of Service Accounts that have `Reader` access to registry          | list(string) | x        |               |
| **writer_serviceaccount**  | List of Service Accounts that have  `Writer`  access to registry        | string       |          | ""            |
| **primary_area**           | Primary area (if multi-region registry)                                 | string       |          | europe        |
| **multi_region**           | Multi-region or single-region registry                                  | bool         |          | true          |
| **public**                 | Is it available for every internet user with `Reader` access? | bool    |          | false         |
| **immutable**              | Enable Immutable tags                                                         | bool         |          | false         |

When you use the GCP private image registry, consider the following:

- The solution is prepared for the GCP Service Account related execution.
- The `roles/artifactregistry.writer` role binding is part of the solution. To learn more, read [Artifact Registry Repository Access Control](https://cloud.google.com/artifact-registry/docs/access-control). If this variable is empty, the solution doesn't add any service account with the `writer` permission.
- The `roles/artifactregistry.reader` role binding is required for lifecycle-manager service accounts. To learn more, read [Artifact Registry Repository Access Control](https://cloud.google.com/artifact-registry/docs/access-control).
- You can make your repository public if you use the `public = true` in the `terraform.tfvars`.
- Vulnerability scanning is enabled by default.