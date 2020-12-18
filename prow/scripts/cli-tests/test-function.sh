#!/usr/bin/env bash

shout "Create local resources for a sample Function"
date
for i in $(seq 1 5); do
    [[ ${i} -gt 1 ]] && echo 'Retrying in 15 seconds..' && sleep 15;
    gcloud compute ssh --quiet --zone="${ZONE}" "${HOST}" -- "sudo kyma init function" && break;
    [[ ${i} -ge 5 ]] && echo "Failed after $i attempts." && exit 1
done;

shout "Apply local resources for the Function to the Kyma cluster"
date
for i in $(seq 1 5); do
    [[ ${i} -gt 1 ]] && echo 'Retrying in 15 seconds..' && sleep 15;
    gcloud compute ssh --quiet --zone="${ZONE}" "${HOST}" -- "sudo kyma apply function" && break;
    [[ ${i} -ge 5 ]] && echo "Failed after $i attempts." && exit 1
done;

sleep 30

shout "Check if the Function is running"
date
attempts=3
for ((i=1; i<=attempts; i++)); do
    set +e
    result=$(gcloud compute ssh --quiet --zone="${ZONE}" "${HOST}" -- "sudo kubectl get pods -lserverless.kyma-project.io/function-name=first-function,serverless.kyma-project.io/resource=deployment -o jsonpath='{.items[0].status.phase}'")
    set -e
    if [[ "$result" == *"Running"* ]]; then
        echo "The Function is in Running state"
        break
    elif [[ "${i}" == "${attempts}" ]]; then
        echo "ERROR: The Function is in ${result} state"
        exit 1
    fi
    echo "Sleep for 15 seconds"
    sleep 15
done
