# Github Statistics

## Overview

`githubstats` fetches statistics for github issues and prints following JSON object:
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

This tool is executed periodically on kyma-workload cluster.
The output is grabbed by StackDriver export (filter is set to "GithubStatsReport" keyword).
JSON object is saved as set of columns in BigQuery database. 
Report can be previewed here: https://datastudio.google.com/s/jlYzET3duNo (access restricted to Kyma developers)

## Usage

To run it, use:
```bash
go run main.go \ 
    --github-access-token={GitHub OAuth2 access token} \
    --github-repo-name={GitHub repository name} \
    --github-repo-owner={GitHub repository owner}
```

### Flags

```bash 
Usage:
  githubstats [flags]

Flags:
  -t, --github-access-token string   GitHub token [Required] [APP_GITHUB_ACCESS_TOKEN]
  -r, --github-repo-name string      repository name [Required] [APP_GITHUB_REPO_NAME]
  -o, --github-repo-owner string     owner/organisation name [Required] [APP_GITHUB_REPO_OWNER]
  -h, --help                         help for githubstats
```


### Environment variables

All flags can be also set using environment variables:

| Name                           | Required | Description                                                           |
| :----------------------------- | :------: | :-------------------------------------------------------------------- |
| **APP_GITHUB_ACCESS_TOKEN**    |    Yes   | The string value with the name of the GitHub OAuth2 access token.     |
| **APP_GITHUB_REPO_OWNER**      |    Yes   | The string value with the name of the organisation/owner.             |
| **APP_GITHUB_REPO_NAME**       |    Yes   | The string value with the name of the repository.                     |
