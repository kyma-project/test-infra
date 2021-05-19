# Github issues


## Overview
This command queries all open Github issues in an organization or repository, and loads that data to a BigQuery table.
### Flags

See the list of available flags:

| Name                             | Required | Description                                                                                          |
| :-----------------------------   | :------: | :--------------------------------------------------------------------------------------------------- |
| **--githubOrgName**              |   Yes    | The string value with the Github organization name to retrieve issues from.
| **--githubRepoName**             |    No    | The string value with the Github repository name to retrieve issues from.
| **--githubToken**                |   Yes    | The string value with the Github OAuth token.
| **--githubBaseURL**              |    No    | The string value with the custom Github API base URL.
| **--issuesFilename**             |    No    | The string value with the name of the generated file with list of issues. It defaults to `issues.json`.
| **--bqCredentials**              |   Yes    | The string value with the path to BigQuery credentials JSON file.
| **--bqProjectID**                |   Yes    | The string value with the name of the BigQuery project.
| **--bqDatasetName**              |   Yes    | The string value with the name of the BigQuery dataset.
| **--bqTableName**                |   Yes    | The string value with the name of the BigQuery table.

## Creating empty table
This program assumes that the table already exists. In order to create new table, do the following:

1. Go to BigQuery console.
2. Create new table in a dataset.
3. Edit the schema as text:
  * For organization copy schema from `table_org_schema.json`
  * For singular repo copy schema from `table_repo_schema.json`
* Select partitioning on `updated_at` field

## Error during upload
Bigquery will fail if the JSON file is bigger than 100MB. The file can be split into smaller parts and uploaded manually before rerunning the program to fix that issue.
