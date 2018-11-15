# Test Infra

## Overview

The purpose of the `test-infra` repository is to store configuration and scripts for the test infrastructure used in the `kyma-project` organization.

### Project structure

<!-- Update the repository structure each time you modify it. -->

The `test-infra` repository has the following structure:

```
  ├── .github                     # Pull request and issue templates             
  ├── development                 # Scripts used for the development of the "test-infra" repository
  ├── docs                        # Documentation for the test infrastructure, such as Prow installation guides
  └── prow                        # Installation scripts for Prow on the production cluster    

```

### Prow

The `test-infra` repository contains the whole configuration of Prow. Its purpose is to replace the internal Continuous Integration (CI) tool in the `kyma-project` organization.

If you want to find out what Prow is, how you can test it, or contribute to it, read the main [`README.md`](./prow/README.md) document in the `prow` folder.

For more detailed documentation, such as installation guides, see the [`docs/prow`](./docs/prow) subfolder.
