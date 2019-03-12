#!/bin/bash

gcloud compute networks describe "$GCLOUD_NETWORK_NAME" 1>/dev/null 2>&1
echo $?