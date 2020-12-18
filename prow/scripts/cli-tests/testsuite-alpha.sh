#!/usr/bin/env bash

shout "Installing Kyma"
date
gcloud compute ssh --quiet --zone="${ZONE}" "${HOST}" -- "yes | sudo kyma alpha deploy --non-interactive ${SOURCE}"

source $TESTDIR/test-version.sh
source $TESTDIR/test-function.sh
source $TESTDIR/test-runtest.sh
