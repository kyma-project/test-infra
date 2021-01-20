#!/usr/bin/env bash

# cli-alpha::deploy starts Kyma installation using the alpha deploy command
#
# Arguments:
#	$1 - OPtional path to an overrides file
function cli-alpha::deploy {
	# shellcheck disable=SC2119
	local overrides=$1

	kyma alpha deploy \
		-v \
		"$( if [ -n "$overrides" ]; then echo "-f $overrides"; fi )"

}