#!/usr/bin/env bash

LIBDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit; pwd)"

# cli-alpha::deploy starts Kyma installation using the alpha deploy command
#
# Arguments:
#	$1 - Path to local resource directory
#	$2 - Path to local components.yaml file
function cli-alpha:deploy {
	kyma alpha deploy \
    	--ci \
    	--resources "${1}" \
    	--components "${2}"
}