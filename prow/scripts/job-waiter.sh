#!/bin/bash

prowJobId=${PROW_JOB_ID}
requiredJobLabelKey=${REQUIRED_JOB_LABEL_KEY:-kyma-job-type}
requiredJobLabelValue=${REQUIRED_JOB_LABEL_VALUE:-component}
maxRetries=${MAX_RETRIES:-100}

eventGUID=$(kubectl get prowjob "${prowJobId}" -ojson | jq -r '.metadata."labels"."event-GUID"')

echo "Prow Job ID: ${prowJobId}"
echo "GUID: ${eventGUID}"
echo "Required label: ${requiredJobLabelKey}=${requiredJobLabelValue}"
echo "Max retries: ${maxRetries}"
echo "===="

i=0
until [[ ${i} -eq ${maxRetries} ]]; do
    unsuccessfulJobs=$(kubectl get prowjob -l event-GUID="${eventGUID}","${requiredJobLabelKey}"="${requiredJobLabelValue}" --no-headers -o=custom-columns='NAME:metadata.name,STATE:status.state' | grep -v 'success')

    if [[ -z "$unsuccessfulJobs" ]]
    then
            echo "No unsuccessful component jobs found. Exiting..."
            exit 0
    fi

    # Check if there are any failed jobs. If so, quit early
    failedJobs=$(echo "${unsuccessfulJobs}" | grep 'error\|failure')

    if [[ -n "$failedJobs" ]];
    then
            echo "Jobs with 'failed' state detected. Exiting with code 1..."
            exit 1
    fi

    echo ">> [${i}/${maxRetries}] Waiting for jobs that should end with success:"
    printf "%s\n" "${unsuccessfulJobs}"

    sleep 5
    ((i++))
done

exit 1