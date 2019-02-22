#!/usr/bin/env bash

# Expected vars:
# - KYMA_PROJECT_DIR - directory of kyma-project sources
# - SNYK_TOKEN - API token used to authenticate in snyk CLI
# - SLACK_TOKEN - Token for Slack bot for which the vulnerabilities reports will be sent

set -e
set -o pipefail

readonly CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"

# authenticate to snyk
snyk auth "${SNYK_TOKEN}"

# test components with snyk
for dir in ${KYMA_SOURCES_DIR}/components/*/
do
  echo "processing ${dir}"
  for file in ${dir}*
  do
    filename=${file##*/} # cut file name
    if [[ ${filename} == "Gopkg.lock" ]]; then
      # fetch dependencies
      echo " ├── fetches dependencies..."
      cd "${dir}"
      dep ensure
      # scan for vulnerabilities
      echo " ├── scanning for vulnerabilities..."
      affectedComponent=${dir%*/} # cut last '/' in dir path
      affectedComponent=${affectedComponent##*/} # cut the path, leave only dir name
      resultsURI=$(snyk monitor --org=kyma-project --project-name="${affectedComponent}" --json | jq -r '.uri')
      # test for high severity vulnerabilities only
      set +e # snyk test return 1 if it find any vulnerabilities, so we need to ignore that
      snyk test --severity-threshold=high --json > snyk-out.json
      set -e
      # send notifications to slack if vulnerabilities was found
      ok=$(jq '.ok' < snyk-out.json)
      if [[ ${ok} == "false" ]]; then
        echo " ├── sending notifications to slack..."
        data='
        {
          "channel": "#kyma-snyk-test",
          "text": "Vulnerabilities of high severity detected!",
          "attachments": [
            {
              "color": "#66ffff",
              "title": "affected component: ",
              "text": "'${affectedComponent}'",
              "actions": [
                {
                  "type": "button",
                  "text": "Show in snyk.io",
                  "url": "'${resultsURI}'"
                }
              ]
            }
          ]
        }'

        for vulnerability in $(jq -c -r '.vulnerabilities | group_by(.id) | map({id:.[0].id,title:.[0].title,packageName:.[0].packageName,semver:.[0].semver}) | .[] | @base64' < snyk-out.json)
        do
          vulnerabilityDecoded=$(printf '%s' "${vulnerability}" | base64 --decode )
          title=$(printf '%s' "${vulnerabilityDecoded}" | jq -r '.title')
          packageName=$(printf '%s' "${vulnerabilityDecoded}" | jq -r '.packageName')
          issueID=$(printf '%s' "${vulnerabilityDecoded}" | jq -r '.id')
          affectedVersions=$(printf '%s' "${vulnerabilityDecoded}" | jq -r '.semver.vulnerable | .[0]')
          newVulnerability='
          {
            "pretext": "Vulnerability: ",
            "color": "#cc3300",
            "title": "'"${title}"'",
            "title_link": "https://snyk.io/vuln/'"${issueID}"'",
            "fields": [
              {
                "title": "Package",
                "value": "'"${packageName}"'",
                "short": true
              },
              {
                "title": "Issue ID",
                "value": "'"${issueID}"'",
                "short": true
              },
              {
                "title": "Affected versions",
                "value": "'"${affectedVersions}"'",
                "short": true
              }
            ]
          }'
          newVulnerability=$(printf '%s' "${newVulnerability}" | jq -r -c ".")
          data=$(echo "${data}" | jq -c '.attachments += ['"${newVulnerability}"']')
        done

        curl -s -X POST \
        -H 'Authorization: Bearer '"${SLACK_TOKEN}" \
        -H 'Content-type: application/json' \
        -H 'cache-control: no-cache' \
        --data "${data}" \
        https://slack.com/api/chat.postMessage \
        > /dev/null # do not show any output

      fi
      echo " └── finished"
    fi
  done
done