#!/usr/bin/env bash

shout "Checking the versions"
clitests::assertRemoteCommand "sudo kyma version"

