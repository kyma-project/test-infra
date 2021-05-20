#!/usr.bin/env bash

set -e

/prow-tools/githubissues \
"--githubOrgName" "kyma-project" \
"--bqProjectID" "sap-kyma-prow" \
"--bqDataset" "github_issues" \
"--bqTable" "issues" \
"--bqCredentials" "${GOOGLE_APPLICATION_CREDENTIALS}" \
"--githubToken" "${BOT_GITHUB_TOKEN}" \


/prow-tools/githubissues \
"--githubOrgName" "kyma" \
"--githubRepoName" "backlog" \
"--githubBaseURL" "https://github.tools.sap/api/v3/" \
"--bqProjectID" "sap-kyma-prow" \
"--bqDataset" "github_issues" \
"--bqTable" "issues_backlog" \
"--bqCredentials" "${GOOGLE_APPLICATION_CREDENTIALS}" \
"--githubToken" "${BOT_GITHUB_TOKEN_}"
