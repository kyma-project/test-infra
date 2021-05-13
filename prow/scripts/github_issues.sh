#!/usr/bin/env bash
gh_local=""

# defiens gh_local token
source "gh_key.sh"

output="issues.json"
#output_labels="labels.json"

BASE_URL="https://api.github.com"
ISSUES_URL="/orgs/kyma-project/issues"
#ISSUES_URL="/search/issues"
params="state=open&filter=all&per_page=70" #&labels=bug"
#params="q=is:open%20org:kyma-project&per_page=100" #&labels=bug"

function prepare_issues() {
    rm "$output"
    # TODO - better loop, response might be limited to 1000 entries
    for i in {1..10}; do
        resp=$(curl -H "Authorization: Bearer $gh_local" "${BASE_URL}${ISSUES_URL}?${params}&page=${i}" 2>/dev/null)
        # echo "$resp" > "$i.json"
        [[ $(echo "$resp" | jq '. | length') == 0 ]] && echo "Finished at page $i, giving $(( i * 90 )) issues at best" && break

        # there are issues in this response, parse themn
        echo "$resp" | jq -c  '.[]' >> "$output"
    done

    echo "Actually there should be around $(wc -l $output) issues in this file"
}



function parse_labels() {

    # $output haver one object per line, let's make proper json table out of this
    #rm "$output"
    # list=()
    # for i in {0..7}; do
    #     # wc -l "$i.json"
    #     list+=("$i.json")
    # done
    # jq -c '.[]' "${list[@]}" > "$output"
    # data=$(cat "$output")
    # echo "$data" | jq -s "." > "$output"

    # now we have all issues in one huge file

    # parse all labels to a flat format, each line contains repo name, issue number and one label
    # if one issue have multiple labels, there will be multiple lines
    jq -c '[. | [. as $in | $in.labels[] as $l | $in | del(.labels) as $in2 | $l * $in2 ] | .[] ] | .[] as $data | {"repository": $data.repository.name, "number": $data.number, "label": $data.name}' $output > $output_labels
    #jq '.repository.name' $output

}

function push_to_bigquery() {
    echo "pushing to bigquery"
    # drop table and recreate it
    bq rm -f -t sap-kyma-prow:test_repos_ph.issues
    # project:dataset.table
    bq load --autodetect --source_format=NEWLINE_DELIMITED_JSON sap-kyma-prow:test_repos_ph.issues "$output"

    # drop table and recreate it
    #bq rm -f -t sap-kyma-prow:test_repos_ph.labels
    #bq load --autodetect --source_format=NEWLINE_DELIMITED_JSON sap-kyma-prow:test_repos_ph.labels "$output_labels"
}

# get all issues and create one big file that bq can understand
prepare_issues

#parse_labels

push_to_bigquery
