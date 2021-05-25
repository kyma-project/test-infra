#!/usr/bin/env bash

set -e

/prow-tools/githubissues \
"--githubOrgName" "kyma-project" \
"--bqProjectID" "sap-kyma-prow" \
"--bqDataset" "github_issues" \
"--bqTable" "github_com_kyma_project" \
"--bqCredentials" "${GOOGLE_APPLICATION_CREDENTIALS}" \
"--githubToken" "${BOT_GITHUB_TOKEN}"


/prow-tools/githubissues \
"--githubOrgName" "kyma" \
"--githubBaseURL" "https://github.tools.sap/api/v3/" \
"--bqProjectID" "sap-kyma-prow" \
"--bqDataset" "github_issues" \
"--bqTable" "github_tools_sap_kyma" \
"--bqCredentials" "${GOOGLE_APPLICATION_CREDENTIALS}" \
"--githubToken" "${BOT_GITHUB_SAP_TOKEN}"
