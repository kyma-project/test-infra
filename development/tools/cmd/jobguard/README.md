# Job Guard 

## Overview

Job Guard is a simple tool that fetches all statuses for GitHub pull requests and waits for some to finish.

The main purpose of the Job Guard is to delay running integration jobs, which depend on components ones. The utility is run as a guard for integration tests.

## Usage

### Run the application

To run the application, run this command:

```bash
COMMIT_SHA={commit_sha} PROW_CONFIG_FILE={prow_config_file} PROW_JOBS_DIRECTORY={prow_jobs_directory} go run main.go
```

### Environment variables

Use the following environment variables to configure the application:

| Name                      | Required  | Default                   | Description |
|---------------------------|-----------|---------------------------|-------------|
| **INITIAL_SLEEP_TIME**    | No        | `1m`                      | The initial sleep time for the application |
| **RETRY_INTERVAL**        | No        | `15s`                     | The interval between re-fetching statuses |
| **TIMEOUT**               | No        | `15m`                     | The timeout of waiting for successful jobs |
| **API_ORIGIN**            | No        | `https://api.github.com`  | The origin of the GitHub API |
| **REPO_OWNER**            | No        | `kyma-project`            | Username or organization name, that owns the repository |
| **REPO_NAME**             | No        | `kyma`                    | The name of the repository |
| **COMMIT_SHA**            | Yes       |                           | The commit SHA |
| **GITHUB_TOKEN**          | No        |                           | The authorization token for GitHub API |
| **JOB_NAME_PATTERN**      | No        | `components`              | The Regexp to filter dependant statuses |
| **PROW_CONFIG_FILE**      | Yes       |                           | The path to the Prow `config.yaml` file  |
| **PROW_JOBS_DIRECTORY**   | Yes       |                           | The path to the directory with Prow jobs |