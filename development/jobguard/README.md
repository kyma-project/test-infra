# JobGuard 

## Overview

JobGuard is a simple tool that fetches all statuses for GitHub pull requests and waits for some of them to finish.
The main purpose of JobGuard is to delay running integration jobs that depend on component jobs. This tool acts as a guard for integration tests.

## Usage

To run the application, use this command:

```shell
go run cmd/jobguard/main.go \
  -github-endpoint=http://ghproxy \
  -github-endpoint=https://api.github.com \
  -github-token-path=/path/to/oauth \
  -org=example-org \
  -repo=example-repo \
  -base-ref=13abc \
  -expected-contexts-regexp="(some-context-regexp|another-context)"
```
## CLI parameters

JobGuard accepts the following command line parameters:

|Flag|Required|Description|
|---|---|---|
|`-github-host`| No | GitHub's default host.|
|`-github-endpoint`| No | GitHub API endpoint.|
|`-github-graphql-endpoint`| No | GitHub GraphQL API endpoint.|
|`-github-token-path`|Yes|Path to the file containing the GitHub OAuth secret.|
|`-debug`|No|Enable debug logging.|
|`-dry-run`|No|Run in dry mode.|
|`-expected-contexts-regexp`|Yes|Regular expression with expected contexts.|
|`-fail-on-no-contexts`|No|Fail if regexp does not match any of the GitHub contexts.|
|`-timeout`|No|Time after JobGuard fails.|
|`-poll-interval`|No|Interval in which JobGuard checks contexts on GitHub.|
|`-org`|Yes|GitHub organisation to check.|
|`-repo`|Yes|GitHub repository to check.|
|`-base-ref`|Yes|GitHub base ref to pull statuses from.|
