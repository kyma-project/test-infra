# Github issues


## Overview

This command queries all open Github issues in an organization or repository, and loads that data to a BigQuery table.

This program assumes that the table already exists. Learn how to [create an empty table](#create-empty-table).
If you get a `no such field` error, [fix the table schema mismatch](#fix-table-schema-mismatch)

If the JSON file is bigger than 100MB, BigQuery fails. To fix that issue, split the file into smaller parts and upload them manually before running the program.

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

## Create empty table

To create an empty table, do the following:

1. Go to BigQuery console.
2. Create a new table in a dataset.
3. Edit the schema as text:
  * For an organization, copy the schema from `table_org_schema.json`.
  * For a singular repo, copy the schema from `table_repo_schema.json`.
4. In the `partitioning` dropdown list, select `updated_at` field.

## Fix table schema mismatch

In order to fix `no such field` error, do the following:

1. Download the current table schema definition. Run: `bq show \ --schema \ --format=prettyjson \ PROJECT_ID:DATASET.TABLE_NAME > table_org_schema.json`.
2. Add missing fields to the `table_org_schema.json` file.
3. Upload the updated table schema. Run:`bq update PROJECT_ID:DATASET.TABLE_NAME table_org_schema.json`.
