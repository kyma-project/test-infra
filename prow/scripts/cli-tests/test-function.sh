#!/usr/bin/env bash

log::info "Create local resources for a sample Function"
clitests::assertRemoteCommand "sudo kyma init function --name first-function --runtime nodejs16"

log::info "Apply local resources for the Function to the Kyma cluster"
clitests::assertRemoteCommand "sudo kyma apply function"

sleep 30

log::info "Check if the Function is running"
clitests::assertRemoteCommand \
    "sudo kubectl get pods -lserverless.kyma-project.io/function-name=first-function,serverless.kyma-project.io/resource=deployment -o jsonpath='{.items[0].status.phase}'" \
    'Running'