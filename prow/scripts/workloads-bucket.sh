#!/usr/bin/env bash

set -e

docker pull alpine:edge
docker tag alpine:edge eu.gcr.io/sap-kyma-prow-workloads/alpine_edge
docker push eu.gcr.io/sap-kyma-prow-workloads/alpine_edge
