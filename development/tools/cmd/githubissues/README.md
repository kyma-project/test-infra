# Github issues


## Overview
This command queries all open Github issues in an organization, and generates BigQuery table containing that data.
### Flags

See the list of available flags:

| Name                             | Required | Description                                                                                          |
| :-----------------------------   | :------: | :--------------------------------------------------------------------------------------------------- |
| **--githubOrgName**              |   Yes    | The string value with the Github organization name to retrieve issues from.
| **--githubToken**                |   Yes    | The string value with the Github OAuth token.
| **--githubBaseURL**              |    No    | The string value with the custom Github API base URL.
| **--issuesFilename**             |    No    | The string value with the name of the generated file with list of issues. It defaults to `issues.json`.
| **--bqCredentials**              |   Yes    | The string value with the path to BigQuery credentials JSON file.
| **--bqProjectID**                |   Yes    | The string value with the name of the BigQuery project.
| **--bqDatasetName**              |   Yes    | The string value with the name of the BigQuery dataset.
| **--bqTableName**                |   Yes    | The string value with the name of the BigQuery table.
