#!/usr/bin/env bash

# cli-alpha::deploy starts Kyma installation using the alpha deploy command
#
# Arguments:
#	$1 - Path to local resource directory
#	$2 - Path to local components.yaml file
function cli-alpha::deploy {
	local overrides=$1

	kyma alpha deploy \
		-v \
		"$( if [ -n "$overrides" ]; then echo "-f $overrides"; fi )"

}