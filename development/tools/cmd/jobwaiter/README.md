# Job Waiter 

## Overview

Job Waiter is a simple tool that fetches all statuses for GitHub pull requests and waits for some to finish.

The main purpose of the Job Waiter is to delay running integration jobs, which depend on components ones. The utility is run as a guard for integration tests.

## Usage

### Run the application

To run the application, run this command:

```bash
PULL_NUMBER={pullNumber} go run main.go
```

Replace values in curly braces with proper details, where:
- `{pullNumber}` is the pull request number.

The service listens on port `3000`.

### Environmental variables

Use the following environment variables to configure the application:

| Name | Required | Default | Description |
|------|----------|---------|-------------|
| **PULL_NUMBER** | Yes | | The pull request number |
| **BOT_GITHUB_TOKEN** | No | | The authorization token for GitHub API|
| **JOB_FILTER_SUBSTRING** | No | | The substring that only dependant job contains in the status name |
| **API_ORIGIN** | No | `https://api.github.com` | The origin of the GitHub API |
| **REPO_OWNER** | No | `kyma-project` | Username or organization name, that owns the repository |
| **REPO_NAME** | No | `kyma` | The name of the repository |
| **INITIAL_SLEEP_TIME** | No | `1m` | The initial sleep time for the application |
| **TICK_TIME** | No | `15s` | The period of statuses re-check |
| **TIMEOUT** | No | `15m` | The timeout of waiting for successful jobs |
