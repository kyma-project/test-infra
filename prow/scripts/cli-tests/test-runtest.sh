#!/usr/bin/env bash

shout "Running a simple test on Kyma"
date
for i in $(seq 1 5); do
    [[ ${i} -gt 1 ]] && echo 'Retrying in 15 seconds..' && sleep 15;
    gcloud compute ssh --quiet --zone="${ZONE}" "${HOST}" -- "sudo kyma test run dex-connection" && break;
    [[ ${i} -ge 5 ]] && echo "Failed after $i attempts." && exit 1
done;

echo "Check if the test succeeds"
date
attempts=3
for ((i=1; i<=attempts; i++)); do
    result=$(gcloud compute ssh --quiet --zone="${ZONE}" "${HOST}" -- "sudo kyma test status -o json" | jq '.status.results[0].status')
    if [[ "$result" == *"Succeeded"* ]]; then
        echo "The test succeeded"
        break
    elif [[ "${i}" == "${attempts}" ]]; then
        echo "ERROR: test result is ${result}"
        exit 1
    fi
    echo "Sleep for 15 seconds"
    sleep 15
done