#!/usr/bin/env bash

log::info "Installing Kyma"
date
gcloud compute ssh --quiet --zone="${ZONE}" "${HOST}" -- "yes | sudo kyma install --non-interactive ${SOURCE}"

# shellcheck disable=SC1090
source "$TESTDIR/test-version.sh"

# shellcheck disable=SC1090
source "$TESTDIR/test-function.sh"

# shellcheck disable=SC1090
source "$TESTDIR/test-runtest.sh"
