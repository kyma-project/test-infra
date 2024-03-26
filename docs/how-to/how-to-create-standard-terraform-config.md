# Standard Terraform Configuration

This document describes the standard Terraform configurationuration that is used in the `test-infra` repository. 

## Infrastructure as Code

The `test-infra` repository uses [Terraform](https://www.terraform.io/) to manage the infrastructure as code. Terraform is an open-source infrastructure as code (IaC) software tool that enables you to safely and predictably create, change, and improve infrastructure. Terraform can manage existing and popular service providers as well as custom in-house solutions.

We decided to build an infrastructure as code because the IaC approach makes the following operations very easy: 
- Reproduce the infrastructure and have a single source of truth for the infrastructure. 
- Test the infrastructure changes before applying them to the production environment.
- Track the changes in the infrastructure and have a history of the changes.
- Revert the changes if needed and maintain documentation of the infrastructure along with the code.

## Terraform Configuration Structure

Our standard structure for Terraform configurationuration is based on the [Google Terraform best practices](https://cloud.google.com/docs/terraform/best-practices-for-terraform) and [Hashicorp Creating Terraform Modules](https://developer.hashicorp.com/terraform/language/modules/develop) articles. Thus, we can easily reuse Terraform modules and share them between different projects. We can also easily test the modules on the development environment before applying them to the production environment.

The Terraform configuration should be stored in the `terraform` directory in the location specific for a use case. See the directory structure for [open policy agent](https://github.com/kyma-project/test-infra/tree/main/opa) in the `test-infra` repository.

```bash

Example structure of the Terraform configuration is the following:

```bash
├── environments
│   ├── dev
│   │   ├── backend.tf
│   │   ├── main.tf
│   │   ├── provider.tf
│   │   ├── terraform.tfvars
│   │   └── variables.tf
│   └── dev2
│       ├── backend.tf
│       ├── main.tf
│       ├── provider.tf
│       ├── terraform.tfvars
│       └── variables.tf
└── modules
    ├── rotate-service-account
    │   ├── main.tf
    │   ├── output.tf
    │   ├── provider.tf
    │   └── variables.tf
    └── service-account-keys-cleaner
        ├── main.tf
        ├── output.tf
        ├── provider.tf
        └── variables.tf
```
### Modules

The `modules` directory contains the Terraform modules that are used in the environments. With modules, we group the resources that are used together, and compose an application component and all needed resources like service account, permissions or messaging system. 

> **Example:** The `rotate-service-account` module contains the resources that are used to rotate the service account keys. The `service-account-keys-cleaner` module contains the resources that are used to clean the service account keys. 

To avoid multiple levels of nesting, a module should not call other modules. Modules should be designed to be called multiple times to create multiple instances of a component. Instances of components should be created as independent entities that can be managed separately. 
For example, if you have two different projects that need to have the same service account, you should create two instances of the `service-account-keys-cleaner` module and pass the project ID as a variable to the module. 

The naming domain should allow creating multiple instances of the same module in the same environment. The module resources naming should allow creating multiple instances of the same module in the same environment. For example, a module should accept a variable value which is used to create a unique name for the modules resources. 

In general, a module should consist of resources unique to the component for which a module was created. Other shared resources like network or storage should be passed to the module as a dependency. The modules directory layout follows the standard Terraform module layout. For more information about Terraform modules layout, refer to the [Terraform standard module structure documentation](https://developer.hashicorp.com/terraform/language/modules/develop/structure).

#### Module Directory Structure

- `main.tf` file contains the definition of resources that are created by the module. Resources definition can be split into multiple files. For example, you can have a `main.tf` file that contains the definition of the common resources that are created by the module and a separate file that contains the definition of the resources composing the cloud-run services that are created by the module.
- `variables.tf` file contains the definition of variables that are passed to the module and used for applying a module. Variables defined in the `variables.tf` file define the module's external API. They must be documented using the **description** attribute. 
- `output.tf` file contains the definition of outputs that are returned by the module. 
- `provider.tf` file contains the definition of the provider that is used by the module.

### Environments

The `environments` directory contains the Terraform configuration for different use cases. The environments separate the infrastructure for different use cases like development, staging and production projects or multiple instances of the same application. 
The name of directories modules defined in `modules` directories are called by the configuration defined in the `environments` directory. Resources and variables specific for environments are passed to the modules as variables values. 
The environments may call modules from any locations, not only the modules defined under the same parent directory. That way, any module existing in `test-infra` and outside of it can be reused. Having a Terraform configuration with the `environments` directory is perfectly fine. Such a configuration simply uses the modules defined in other locations and provides the definition of resources specific to the use case. Outputs returned by the environments are published to the Terraform remote state. It's recommended to output all resources from an environment so that other environments can consume it and use it as a dependency.

The Terraform configuration for our production environment is stored in the `test-infra` repository in this [location](https://github.com/kyma-project/test-infra/tree/main/configurations/terraform/environments/prod). This is the root module for production usage. Maintaining only one root module for production usage makes configuration simpler and reduces dependencies to other modules, limiting it to the calls to the modules defined in the `modules` directory. It also lets you avoid duplication in the Terraform configuration. This simplifies maintaining resource and data definitions that are specific to the production environment but don't belong to any module specifically.

The Terraform configuration for the development environment is usually stored in the same `terraform` directory as the module. This architecture allows us to easily test the modules on the development environment independent of the production environment.

#### Environment Directory Structure

- `main.tf` file contains the definition of resources that are created by the environment. Resources definition can be split into multiple files. Calls to the modules should be defined in these files. Outputs returned by the environments can be defined in these files instead of `output.tf` file.
- `variables.tf` file contains the definition of variables that are passed to the environment and used for applying the environment. Variables defined in the `variables.tf` file define the environment's external API. Variables defined in `variables.tf` file must be documented using the description attribute. Values of the variables defined in `variables.tf` file should be provided in the `terraform.tfvars` file.
- `backend.tf` file contains the definition of the backend that is used by the module. Terraform modules must use Google Cloud (gcp) as a remote state storage.
- `terraform.tfvars` file is used to define the values of the variables that are used by the environment. The `terraform.tfvars` file must be created for each environment and stored in the environment directory. The path to the `terraform.tfvars` file is passed to the `terraform` CLI command.
- `provider.tf` file contains the definition of the provider that is used by the environment.

## Terraform Configuration Usage

The Terraform configuration must be tested and applied automatically through our CI/CD pipeline. 

For testing the configuration, we use a presubmit ProwJob that runs the `terraform plan` command and checks if the plan is valid. For applying the configuration, we use a postsubmit ProwJob that runs the `terraform apply` command. 
Both ProwJobs use the same remote state to make sure `terraform plan` is executed on the same state as `terraform apply`. Moreover, GCP remote state supports remote locking. Remote locking ensures that only one `terraform apply` is executed at the same time and our systems are in consistent state.

ProwJobs applying the Terraform configuration use a Terraform executor image that contains a Terraform CLI and the helper tool [tfcmt](https://suzuki-shunsuke.github.io/tfcmt/), which adds comments to the GitHub pull request (PR) with the Terraform plan output. This makes it easier to review the results of Terraform actions. 

Usually, Terraform executor ProwJobs are executed on every change in the Terraform configuration. The ProwJobs are executed only for the Terraform configuration that was changed. Running Terraform executor on changes in other files may be needed. For example, changes in [workflow definition file](https://github.com/kyma-project/test-infra/blob/main/pkg/gcp/workflows/secrets-leak-detector.yaml) require running Terraform executor to reflect changes in respective environments.

### Terraform Presubmit ProwJob

Here's an example of the [presubmit ProwJob](https://github.com/kyma-project/test-infra/blob/4540c0ba3622b4f1fed47a50dedc189fdfc324b1/prow/jobs/test-infra/secrets-rotator.yaml) for the secrets-rotator application:
- The presubmit ProwJob runs the `terraform plan` command and publishes the results on a GitHub PR.

Here's an example of the [postubmit ProwJob](https://github.com/kyma-project/test-infra/blob/4540c0ba3622b4f1fed47a50dedc189fdfc324b1/prow/jobs/test-infra/secrets-rotator.yaml) for the secrets-rotator application:
- The postsubmit ProwJob runs the `terraform apply` command and publishes the results on a GitHub PR.
