# Github Release

## Overview

This command enables Github authentication on Kyma. To achieve that, 2 source files are edited: `dex-config-map.yaml` and `cluster-users/values.yaml`.



### Environment variables

Available environment variables:

| Name                                         | Required | Description                                                                                          |
| :------------------------------------------- | :------: | :--------------------------------------------------------------------------------------------------- |
| **KYMA_PROJECT_DIR**                         |    Yes   | Path to the kyma-project directory which contains Kyma source code  |
| **GITHUB_INTEGRATION_APP_CLIENT_ID**         |    Yes   | Client ID of Github application  |
| **GITHUB_INTEGRATION_APP_CLIENT_SECRET**     |    Yes   | Client secret of Github application  |
| **DEX_CALLBACK_URL**                         |    Yes   | DEX callback URL |
| **GITHUB_TEAMS_WITH_KYMA_ADMINS_RIGHTS**     |    Yes   | Comma-separated list of Github teams in kyma-project that are bound to `kymaAdmins` role |
