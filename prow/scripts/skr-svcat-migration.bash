#!/usr/bin/env bash

# Script to run SKR svcat migration

set -euo pipefail
make -C /home/prow/go/src/github.com/kyma-project/kyma/tests/fast-integration ci-skr-svcat-migration
