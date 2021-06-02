#!/usr/bin/env bash




# clusterProvisioner::generateCommonName create and export COMMON_NAME variable
# it generates random part of COMMON_NAME and prefix it with provided arguments
#
# Arguments:
# $1 - string to use as a common name prefix /optional
# $2 - pull request number or commit id to use as a common name prefix /optional
#clusterProvisioner::generateCommonName() {
#  NAME_PREFIX=$1
#  PULL_NUMBER=$2
#  if [ ${#PULL_NUMBER} -gt 0 ]; then
#    PULL_NUMBER="-${PULL_NUMBER}-"
#  fi
#  RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c6)
#  COMMON_NAME=$(echo "${NAME_PREFIX}${PULL_NUMBER}${RANDOM_NAME_SUFFIX}" | tr "[:upper:]" "[:lower:]")
#  export COMMON_NAME
#}


# clusterProvisioner::generateCommonName create and export COMMON_NAME variable
# it generates random part of COMMON_NAME and prefix it with provided arguments
#
# Arguments:
# $1 - string to use as a common name prefix /optional
# $2 - pull request number or commit id to use as a common name prefix /optional
clusterProvisioner::provision_cluster() {
  clusterName=$1
  credentials=$2
  gcpProject=$3
  location=$4
  machineType=$5
  k8sVersion=$6

  log::info "Provision cluster: \"${clusterName}\""

  export CLEANUP_CLUSTER="true"
  (
  kyma provision gke \
    --ci \
    --non-interactive \
    --verbose \
    --name "${clusterName}" \
    --project "${gcpProject}" \
    --credentials "${credentials}" \
    --location "${location}" \
    --type "${machineType}" \
    --kube-version="${k8sVersion}"
  )

  if [ "${DEBUG_COMMANDO_OOM}" = "true" ]; then
    # run oom debug pod
    utils::debug_oom
  fi
}
