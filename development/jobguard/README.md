# Job Guard 

## Overview

Job Guard is a simple tool that fetches all statuses for GitHub pull requests and waits for some of them to finish.
The main purpose of Job Guard is to delay running integration jobs that depend on component jobs. This tool acts as a guard for integration tests.

## Usage

### Run the application

To run the application, use this command:

```bash
COMMIT_SHA={commit_sha} PROW_CONFIG_FILE={prow_config_file} PROW_JOBS_DIRECTORY={prow_jobs_directory} go run cmd/main.go
```

### Environment variables

Use the following environment variables to configure the application:

| Name                      | Required  | Default  value                 | Description |
|---------------------------|:-----------:|---------------------------|-------------|
| **INITIAL_SLEEP_TIME**    | NO        | `1m`                      | Initial sleep time for the application |
| **RETRY_INTERVAL**        | NO        | `15s`                     | Interval between re-fetching statuses |
| **TIMEOUT**               | NO        | `15m`                     | Timeout of waiting for successful jobs |
| **API_ORIGIN**            | NO        | `https://api.github.com`  | Origin of the GitHub API |
| **REPO_OWNER**            | NO        | `kyma-project`            | Username or organization name of the repository owner |
| **REPO_NAME**             | NO        | `kyma`                    | Name of the repository |
| **COMMIT_SHA**            | YES       |                           | Commit SHA |
| **GITHUB_TOKEN**          | NO        |                           | Authorization token for GitHub API |
| **JOB_NAME_PATTERN**      | NO        | `components`              | Regexp to filter dependant statuses |
| **PROW_CONFIG_FILE**      | YES       |                           | Path to the Prow `config.yaml` file  |
| **PROW_JOBS_DIRECTORY**   | YES       |                           | Path to the directory with Prow jobs |
