#!/usr/bin/env bash

# Description: Applies the coreDNS patch to a k3s cluster on gcloud
#

export REGISTRY_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' /registry.localhost) 
kubectl -n kube-system patch cm coredns --patch "$(cat ~/test-infra/prow/scripts/resources/k3d-coredns-patch.tpl.yaml | envsubst )"
