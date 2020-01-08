#!/usr/bin/env bash

# Expected vars:
# - KYMA_PROJECT_DIR - directory of kyma-project sources
# - SNYK_TOKEN - API token used to authenticate in snyk CLI
# - SAP_SLACK_BOT_TOKEN - Token for Slack bot for which the vulnerabilities reports will be sent

set -e
set -o pipefail

discoverUnsetVar=false

for var in KYMA_PROJECT_DIR SNYK_TOKEN SAP_SLACK_BOT_TOKEN; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

readonly CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"

function authenticate() {
  snyk auth "${SNYK_TOKEN}"
}

function sendSlackNotification() {
  AFFECTED_COMPONENT="$1"
  SNYK_URI="$2"
  SLACK_CHANNEL="#kyma-snyk-test"

  DATA='
  {
    "channel": "'${SLACK_CHANNEL}'",
    "text": "Vulnerabilities of high severity detected!",
    "attachments": [
      {
        "color": "#66ffff",
        "title": "affected component: ",
        "text": "'${AFFECTED_COMPONENT}'",
        "actions": [
          {
            "type": "button",
            "text": "Show in snyk.io",
            "url": "'${SNYK_URI}'"
          }
        ]
      }
    ]
  }'

  VULNERABILITIES=$(jq -c -r '.vulnerabilities | group_by(.id) | map({id:.[0].id,title:.[0].title,packageName:.[0].packageName,semver:.[0].semver}) | .[] | @base64' < snyk-out.json)
  for VULNERABILITY in ${VULNERABILITIES}
  do
    VULNERABILITY_DECODED=$(printf '%s' "${VULNERABILITY}" | base64 --decode )
    TITLE=$(printf '%s' "${VULNERABILITY_DECODED}" | jq -r '.title')
    PACKAGE_NAME=$(printf '%s' "${VULNERABILITY_DECODED}" | jq -r '.packageName')
    ISSUE_ID=$(printf '%s' "${VULNERABILITY_DECODED}" | jq -r '.id')
    AFFECTED_VERSIONS=$(printf '%s' "${VULNERABILITY_DECODED}" | jq -r '.semver.vulnerable | .[0]')
    NEWV_ULNERABILITY='
    {
      "pretext": "Vulnerability: ",
      "color": "#cc3300",
      "title": "'"${TITLE}"'",
      "title_link": "https://snyk.io/vuln/'"${ISSUE_ID}"'",
      "fields": [
        {
          "title": "Package",
          "value": "'"${PACKAGE_NAME}"'",
          "short": true
        },
        {
          "title": "Issue ID",
          "value": "'"${ISSUE_ID}"'",
          "short": true
        },
        {
          "title": "Affected versions",
          "value": "'"${AFFECTED_VERSIONS}"'",
          "short": true
        }
      ]
    }'
    NEWV_ULNERABILITY=$(printf '%s' "${NEWV_ULNERABILITY}" | jq -r -c ".")
    DATA=$(echo "${DATA}" | jq -c '.attachments += ['"${NEWV_ULNERABILITY}"']')
  done

  curl -s -X POST \
  -H 'Authorization: Bearer '"${SAP_SLACK_BOT_TOKEN}" \
  -H 'Content-type: application/json' \
  -H 'cache-control: no-cache' \
  --data "${DATA}" \
  https://slack.com/api/chat.postMessage \
  > /dev/null # do not show any output

}

function testComponents() {
  for DIR in ${KYMA_SOURCES_DIR}/components/*/
  do
    TESTED_COMPONENT=$(basename "${DIR}")
    KYMA_PROJECT="kyma-project"
    echo "processing ${DIR}"

    GOPKG_FILE_NAME="${DIR}"Gopkg.lock

    if [ -f "${GOPKG_FILE_NAME}" ]; then
      # fetch dependencies
      echo " ├── fetching dependencies..."
      cd "${DIR}"
      dep ensure --vendor-only

      # scan for vulnerabilities
      echo " ├── scanning for vulnerabilities..."
      set +e
      snyk test --severity-threshold=high --json > snyk-out.json
      
      # check if snyk project is 
      echo " ├── sending snyk report..."
      SNYK_MONITOR_STATUS=$(snyk monitor --org="${KYMA_PROJECT}" --project-name="${TESTED_COMPONENT}" --json)
      set -e
      
      if [[ $(echo "$SNYK_MONITOR_STATUS" | jq -r '.ok') == "false" ]]; then
        echo "$SNYK_MONITOR_STATUS" | jq -r '.error'
        echo "There was an error with Snyk monitor command. Check above response for more information."
        exit 1
      else
        PROJECT_URI=$(echo "$SNYK_MONITOR_STATUS" | jq -r '.uri')
      fi
      # send notifications to slack if vulnerabilities was found
      OK=$(jq '.ok' < snyk-out.json)
      if [[ ${OK} == "false" ]]; then
        echo " ├── sending notifications to slack..."

        # COMPONENT_TO_TEST=$(basename "${DIR}")
        sendSlackNotification "${TESTED_COMPONENT}" "${PROJECT_URI}"
      else
        echo " ├── No vulnerabilities found."
      fi
      
    else
      echo " ├── No Gopkg.lock found"
    fi
    echo " └── finished"
  done
}

# authenticate to snyk
authenticate

# test components with snyk
testComponents
