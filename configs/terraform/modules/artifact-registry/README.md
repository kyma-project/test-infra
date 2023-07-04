# gcp-artifact-registry-terraform

Provide private image registry in GCP

## Usage

Configure variables in `terraform.tfvars` (configs/terraform/environments/prod/terraform.tfvars) file

Default values

```terraform
artifact_registry_module         = "cap-operator"
artifact_registry_owner          = "neighbors"
artifact_registry_type           = "development"
immutable_artifact_registry      = false
artifact_registry_multi_region   = true
artifact_registry_serviceaccount = "sa-kyma-project@sap-kyma-prow.iam.gserviceaccount.com"
```

Note:
- Soluton is prepared for GCP Service Account related execution
- Credential file is the service account key file in json format
- `roles/artifactregistry.write` role binding is part of the solution ([Artifact Registry Repository Access Control](https://cloud.google.com/artifact-registry/docs/access-control))
- Vulnerability scanning is on by default


