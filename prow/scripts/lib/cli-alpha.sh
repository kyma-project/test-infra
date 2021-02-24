#!/usr/bin/env bash

# cli-alpha::deploy starts Kyma installation using the alpha deploy command
#
function cli-alpha::deploy {
  if [[ "$EXECUTION_PROFILE" == "evaluation" ]]; then
	  kyma alpha deploy --ci --profile evaluation --verbose
	else
	  kyma alpha deploy --ci --verbose
	fi
}