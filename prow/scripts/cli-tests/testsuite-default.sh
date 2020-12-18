#!/usr/bin/env bash

shout "Installing Kyma"
date
gcloud compute ssh --quiet --zone="${ZONE}" "${HOST}" -- "yes | sudo kyma install --non-interactive ${SOURCE}"

source $TESTDIR/test-version.sh
source $TESTDIR/test-function.sh
source $TESTDIR/test-runtest.sh
