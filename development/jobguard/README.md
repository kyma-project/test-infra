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
| **INITIAL_SLEEP_TIME**    | No        | `1m`                      | Initial sleep time for the application |
| **RETRY_INTERVAL**        | No        | `15s`                     | Interval between re-fetching statuses |
| **TIMEOUT**               | No        | `15m`                     | Timeout of waiting for successful jobs |
| **API_ORIGIN**            | No        | `https://api.github.com`  | Origin of the GitHub API |
| **REPO_OWNER**            | No        | `kyma-project`            | Username or organization name of the repository owner |
| **REPO_NAME**             | No        | `kyma`                    | Name of the repository |
| **COMMIT_SHA**            | Yes       | None                          | Commit SHA |
| **GITHUB_TOKEN**          | No        | None                          | Authorization token for GitHub API |
| **JOB_NAME_PATTERN**      | No        | `components`              | Regexp to filter dependant statuses |
| **PROW_CONFIG_FILE**      | Yes       | None                          | Path to the Prow `config.yaml` file  |
| **PROW_JOBS_DIRECTORY**   | Yes       | None                          | Path to the directory with Prow jobs |
