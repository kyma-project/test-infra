# gcp-artifact-registry-terraform

Provide private image registry in GCP

## Usage

Configure variables in `terraform.tfvars` (configs/terraform/environments/prod/terraform.tfvars) file

Default values

```terraform
module                         = "cap-operator"
type                           = "development"
immutable_artifact_registry    = false
artifact_registry_multi_region = true
```

Note:
- Soluton is prepared for GCP Service Account related execution
- Credential file is the service account key file in json format
- `roles/artifactregistry.write` role binding is part of the solution ([Artifact Registry Repository Access Control](https://cloud.google.com/artifact-registry/docs/access-control))
- Vulnerability scanning is on by default


