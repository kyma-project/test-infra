#!/usr/bin/env bash

kubectl create configmap config --from-file=config.yaml=config.yaml --dry-run -o yaml | kubectl replace configmap config -f -