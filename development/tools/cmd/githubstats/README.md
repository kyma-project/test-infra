# GitHub Statistics

## Overview

`githubstats` fetches statistics for GitHub issues and prints the following JSON object:
```json
{
  "Issues": {
    "Open": {
      "TotalCount": 391,
      "Bugs": 497,
      "PriorityCritical": 128,
      "Regressions": 0,
      "TestFailing": 99,
      "TestMissing": 34
    },
    "Closed": {
      "TotalCount": 2228,
      "Bugs": 497,
      "PriorityCritical": 128,
      "Regressions": 0,
      "TestFailing": 99,
      "TestMissing": 34
    }
  },
  "Type": "GithubStatsReport",
  "Owner": "kyma-project",
  "Repository": "kyma",
  "Timestamp": "2019-12-09T09:37:29.2027749Z"
}
```

This tool is executed periodically on the `kyma-workload` cluster.
The output is grabbed by a StackDriver export (the filter is set to the "GithubStatsReport" keyword).
JSON object is saved as a set of columns in the BigQuery database. 
The report preview is available [here](https://datastudio.google.com/s/jlYzET3duNo) (the access is restricted to the Kyma developers).

## Usage

To run it, use:
```bash
go run main.go \ 
    --github-access-token={GitHub OAuth2 access token} \
    --github-repo-name={GitHub repository name} \
    --github-repo-owner={GitHub repository owner}
```

### Flags

Usage:
```bash
  githubstats [flags]
```

Flags:
```bash
  -t, --github-access-token string   GitHub token [Required] [APP_GITHUB_ACCESS_TOKEN]
  -r, --github-repo-name string      repository name [Required] [APP_GITHUB_REPO_NAME]
  -o, --github-repo-owner string     owner/organisation name [Required] [APP_GITHUB_REPO_OWNER]
  -h, --help                         help for githubstats
```


### Environment variables

All flags can also be set using the environment variables:

| Name                           | Required | Description                                                           |
| :----------------------------- | :------: | :-------------------------------------------------------------------- |
| **APP_GITHUB_ACCESS_TOKEN**    |    Yes   | The string value with the name of the GitHub OAuth2 access token.     |
| **APP_GITHUB_REPO_OWNER**      |    Yes   | The string value with the name of the organization/owner.             |
| **APP_GITHUB_REPO_NAME**       |    Yes   | The string value with the name of the repository.                     |
