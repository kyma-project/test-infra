# Standard Terraform Config

This document describes the standard Terraform configuration that is used in test-infra repository. 

## Infrastructure as Code

The test-infra repository uses [Terraform](https://www.terraform.io/) to manage the infrastructure as code. Terraform is an open-source infrastructure as code software tool that enables you to safely and predictably create, change, and improve infrastructure. Terraform can manage existing and popular service providers as well as custom in-house solutions. We made a decision to build an infrastructure as code to be able to easily reproduce the infrastructure and to have a single source of truth for the infrastructure. IaC approach also allows us to easily test the infrastructure changes before applying them to the production environment. Additionally, it allows us to easily track the changes in the infrastructure and to have a history of the changes. It allows us to easily revert the changes if needed and maintain documentation of the infrastructure along with the code.

## Terraform config structure

Our standard structure for Terraform config is based on the [Google Terraform best practices](https://cloud.google.com/docs/terraform/best-practices-for-terraform) and [Hashicorp Creating Terraform Modules](https://developer.hashicorp.com/terraform/language/modules/develop) articles. It allows us easily reuse terraform modules and share them between different projects. It also allows us to easily test the modules on development environment before applying them to the production environment. Terraform config should be stored in the `terraform` directory in the location specific for use case. See directory structure for [secrets-rotator](https://github.com/kyma-project/test-infra/tree/main/development/secrets-rotator) application in test-infra repository.

```bash

Example structure of the Terraform config is the following:

```bash
├── environments
│   ├── dev
│   │   ├── backend.tf
│   │   ├── main.tf
│   │   ├── provider.tf
│   │   ├── terraform.tfvars
│   │   └── variables.tf
│   └── prod
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

The `modules` directory contains the Terraform modules that are used in the environments. The modules are used to group the resources that are used together and compose an application component and all needed resources like service account, permissions or messaging system. For example, the `rotate-service-account` module contains the resources that are used to rotate the service account keys. The `service-account-keys-cleaner` module contains the resources that are used to clean the service account keys. A module should not call other modules, to avoid multiple levels of nesting. Modules should be designed to be called multiple times to create multiple instances of a component. Instances of components should be created as independent entities that can be managed separately. For example, if you have two different projects that need to have the same service account, you should create two instances of the `service-account-keys-cleaner` module and pass the project id as a variable to the module. The naming domain should allow creating multiple instances of the same module in the same environment. The module resources naming should allow creating multiple instances of the same module in the same environment. For example a module should accept a variable value which is used to create a unique name for the modules resources. In general a modul should consist of resources uniq to the component for which a module was created. Other shared resources like network or storage should be passed to the module as a dependency. Modules directory layout is following standard Terraform module layout. For more information about Terraform modules layout please refer to the [Terraform standard module structure documentation](https://developer.hashicorp.com/terraform/language/modules/develop/structure).

#### Module directory structure

- `main.tf` file contains definition of resources that are created by the module. Resources definition can be split into multiple files. For example, you can have a `main.tf` file that contains definition of the common resources that are created by the module and a separate file that contains definition of the resources composing cloud run services that are created by the module.
- `variables.tf` file contains definition of variables that are passed to the module and used for applying a module. Variables defined in `variables.tf` file define module external API. Variables defined in `variables.tf` file must be documented using the description attribute. 
- `output.tf` file contains definition of outputs that are returned by the module. 
- `provider.tf` file contains definition of the provider that is used by the module.

### Environments

The `environments` directory contains the Terraform config for different use cases. The environments are used to separate the infrastructure for different use cases like development, staging and production projects or multiple instances of the same application. The name of directories Modules defined in `modules` directories are called by config defined in environments directory. Resources and variables specific for environments are passed to the modules as a variables values. Environments may call modules from any locations not only modules defined under the same parent directory. That way any module existing in test-infra and outside of it, can be reused. It's perfectly fine to have a terraform config with `environments` directory only. Such config simply use modules defined in other locations and provide definition of resources specific for use case. Outputs returned by environments are published to the Terraform remote state. It's preferred to output all resources from environment. This allows other environments consume it and use it as a dependency. 

#### Environment directory structure

- `main.tf` file contains definition of resources that are created by the environment. Resources definition can be split into multiple files. Calls to the modules should be defined in the these files. Outputs returned by the environments can be defined in these files instead of `output.tf` file.
- `variables.tf` file contains definition of variables that are passed to the environment and used for applying the environment. Variables defined in `variables.tf` file define environment external API. Variables defined in `variables.tf` file must be documented using the description attribute. Values of the variables defined in `variables.tf` file should be provided in the `terraform.tfvars` file.
- `backend.tf` file contains definition of the backend that is used by the module. Terraform modules must use Google gcp as a remote state storage.
- `terraform.tfvars` file is used to define the values of the variables that are used by the environment. The `terraform.tfvars` file should be created for each environment and should be stored in the environment directory. The path to the `terraform.tfvars` file is passed to the terraform cli command.
- `provider.tf` file contains definition of the provider that is used by the environment.

## Terraform config usage

Terraform config must be tested and applied in an automated way through our CI/CD pipeline. For testing config we use a presubmit prowjob which runs terraform plan command and checks if the plan is valid. For applying config we use a postsubmit prowjob which runs terraform apply command. Both prowjobs use the same remote state to make sure Terraform plan is executed on the same state as Terraform apply. Moreover, a Google GCP remote state support remote locking. Remote locking is used to make sure that only one Terraform apply is executed at the same time and our systems are in consistent state. Prowjobs applying Terraform config are using a terraform executor image contains a terraform cli and a helper tool tfcmt. A tool adds comments to the GitHub pull request with the Terraform plan output. This makes easier to review the terraform actions results. Usualy terraform executor prowjobs are executed on every change in the Terraform config. The prowjobs are executed only for the Terraform config that was changed. It may be needed to run terraform executor on changes in other files. For example, changes in [workflow definition file](https://github.com/kyma-project/test-infra/blob/main/development/gcp/workflows/secrets-leak-detector.yaml) require running terraform executor to reflect changes in respective environemnts.

### Terraform presubmit prowjob

Example of the [presubmit prowjob](https://github.com/kyma-project/test-infra/blob/4540c0ba3622b4f1fed47a50dedc189fdfc324b1/prow/jobs/test-infra/secrets-rotator.yaml#L92) for the secrets-rotator application. Presubmit runs terraform plan command and publish results on a GitHub pull request.

Example of the [postubmit prowjob](https://github.com/kyma-project/test-infra/blob/4540c0ba3622b4f1fed47a50dedc189fdfc324b1/prow/jobs/test-infra/secrets-rotator.yaml#L222) for the secrets-rotator application. A Postsubmit runs terraform apply command and publish results on a GitHub pull request.
