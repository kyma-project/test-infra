#!/usr/bin/env bash

log::info "Checking the versions"
clitests::assertRemoteCommand "sudo kyma version"
