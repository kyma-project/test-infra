#!/usr/bin/env bash
LIBDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit; pwd)"

# shellcheck source=prow/scripts/lib/log.sh
source "${LIBDIR}/log.sh"

## Scan for logs in KMC
function kmc::test {
  log::info "###Testing KMC###"
  kubectl get po -n kcp-system -l app=kyma-metrics-collector
  kubectl logs -n kcp-system -l app=kyma-metrics-collector
}
## Search for a particular pattern of metrics sent to EDP

## If found the job is successful

## Else job failed