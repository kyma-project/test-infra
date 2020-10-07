# Enable Github Auth

## Overview

This command enables Github authentication on Kyma. Internally, it updates these two source files: 
- `dex-config-map.yaml` - adds the  Github connector to Dex
- `cluster-users/values.yaml` - binds Github groups to the default roles added to every Kyma Namespace


### Environment variables

These are the available environment variables:

| Name                                         | Required | Description                                                                                          |
| :------------------------------------------- | :------: | :--------------------------------------------------------------------------------------------------- |
| **KYMA_PROJECT_DIR**                         |    Yes   | Path to the `kyma-project` directory which contains the Kyma source code  |
| **GITHUB_INTEGRATION_APP_CLIENT_ID**         |    Yes   | Client ID of an OAuth Github application  |
| **GITHUB_INTEGRATION_APP_CLIENT_SECRET**     |    Yes   | Client secret of an OAuth Github application  |
| **DEX_CALLBACK_URL**                         |    Yes   | DEX callback URL |
| **GITHUB_TEAMS_WITH_KYMA_ADMINS_RIGHTS**     |    Yes   | Comma-separated list of Github teams in `kyma-project` that are bound to the `kymaAdmin` role |

## More
Dex docs for GitHub connector: https://dexidp.io/docs/connectors/github/
