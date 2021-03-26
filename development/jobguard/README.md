# Job Guard 

## Overview

Job Guard is a simple tool that fetches all statuses for GitHub pull requests and waits for some of them to finish.
The main purpose of Job Guard is to delay running integration jobs that depend on component jobs. This tool acts as a guard for integration tests.

## Usage

### Run the application

To run the application, use this command:

|Flag|Required|Description|
|---|---|---|
|`-github-host`| No | GitHub's default host |
|`-github-endpoint`| No | GitHub API endpoint|
|`-github-graphql-endpoint`| No | GitHub GraphQL API endpoint|
|`-github-token-path`|Yes|Path to the file containing the GitHub OAuth secret|
|`-debug`|No|Enable debug logging|
|`-dry-run`|No|Run in dry mode|

