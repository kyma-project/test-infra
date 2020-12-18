#!/usr/bin/env bash

shout "Checking the versions"
date
for i in $(seq 1 5); do
    [[ ${i} -gt 1 ]] && echo 'Retrying in 15 seconds..' && sleep 15;
    gcloud compute ssh --quiet --zone="${ZONE}" "${HOST}" -- "sudo kyma version" && break;
    [[ ${i} -ge 5 ]] && echo "Failed after $i attempts." && exit 1
done;
