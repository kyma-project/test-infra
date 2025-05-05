# IaC Configuration Guideline

This document outlines the standard Terraform configuration and provides guidelines.

## Infrastructure as Code

The `repository uses `[Terraform](https://www.terraform.io/) to manage infrastructure as code (IaC). Terraform provides a predictable,
safe, and efficient approach to creating, modifying, and managing infrastructure. IaC simplifies:

- Reproducibility of infrastructure through a single source of truth.
- Pre-validation of infrastructure changes before deployment.
- Infrastructure change tracking and history.
- Easy rollback and improved documentation through code integration.

## IaC Configuration Structure

Our configuration aligns with [Google Terraform best practices](https://cloud.google.com/docs/terraform/best-practices-for-terraform)
and [Hashicorp module development guidelines](https://developer.hashicorp.com/terraform/language/modules/develop), facilitating module reuse
and streamlined testing.

Terraform configurations must reside within the `terraform` directory, structured and named per application or use case. Grouping by
resource type is not permitted.

Example structure:

```bash
terraform
├── environments
│   └── prod
│       ├── backend.tf
│       ├── main.tf
│       ├── use-case-1.tf
│       ├── scanner-app.tf
│       ├── provider.tf
│       └── variables.tf
└── modules
    ├── application-one
    │   ├── main.tf
    │   ├── output.tf
    │   ├── provider.tf
    │   └── variables.tf
    └── application-two
        ├── main.tf
        ├── output.tf
        ├── provider.tf
        └── variables.tf
```

For better readability and maintenance, the in config definitions must always be preferred over cli parameters or environment variables.

When creating multiple instances of any resource—whether directly or via a module—each instance must be defined explicitly in the Terraform
configuration.
Avoid using `for_each`, `count`, or passing lists of values to generate multiple resources.
This improves readability, simplifies change tracking, and reduces complexity in infrastructure code.

**Example**

Instead of:

```hcl
resource "google_project_iam_member" "example" {
  for_each = toset(["user:a@example.com", "user:b@example.com"])
  project = var.project_id
  role    = "roles/viewer"
  member  = each.value
}
```

Use:

```hcl
resource "google_project_iam_member" "user_a" {
  project = var.project_id
  role    = "roles/viewer"
  member  = "user:a@example.com"
}

resource "google_project_iam_member" "user_b" {
  project = var.project_id
  role    = "roles/viewer"
  member  = "user:b@example.com"
}
```

### Modules

Modules group and encapsulate resources specific to applications or use cases.
These grouped resources compose application components and all needed related resources like service account, permissions, or messaging
system.
Modules must be preferred over raw resource definitions in the environment configuration files.

Modules:

- Must allow independent use in multiple instances and environments (including dev projects).
- Accept parameters for creating unique resource names.
- Use input variables for dependencies or resource sharing (network, storage).

**Example Module Directory Structure**

- `main.tf`: Resource definitions. Can split definitions across multiple logical files.
- `variables.tf`: Input variables defining the external API of the module. Documented using the `description` attribute.
- `output.tf`: Outputs returned by the module.
- `provider.tf`: Providers utilized by the module.

### Environments

The `environments` directory contains configurations for distinct environments (e.g., production).
Each environment configuration is located in a separate directory.
Configurations within a directory representing an environment, like `environments/prod`, can call modules from any location, even external
sources.

**Example Environment Directory Structure**

- `main.tf`: Environment-specific resource and module calls.
- `variables.tf`: Environment input variables. Documented with `description`.
- `backend.tf`: Backend configuration, typically using Google Cloud Storage (GCS).
- `main.tf`: Environment-specific resource and module calls.
- `use-case-1.tf`: Resources or modules calls related to a specific use case.
- `scanner-app.tf`: Resources or modules calls related to the scanner application.

## IaC Configuration Files Format

- `.tfvars` files are not used to provide variables values.
- Input values are explicitly provided via variable defaults.
- Modules and resources must receive variables or other resources as input. It’s not allowed to pass string literals.
- Environments must prefer to utilize modules for resource creation.

## IaC Configuration Usage

IaC configuration is tested and applied via the CI/CD pipeline:

- Pull request pipeline runs `terraform plan`, verifying changes, and publishing results on GitHub PR.
- Push pipeline runs `terraform apply`, applying changes, and posting results on GitHub PR.

Both jobs utilize a shared remote state (GCP backend), ensuring consistent, locked states for safe concurrent execution.

Executor images for Terraform jobs include Tofu CLI and [tfcmt](https://suzuki-shunsuke.github.io/tfcmt/) for detailed GitHub PR
annotations.

Terraform jobs trigger automatically upon configuration file modifications. Changes in associated files (e.g., workflows) may also trigger
Terraform execution to reflect updated configurations.

