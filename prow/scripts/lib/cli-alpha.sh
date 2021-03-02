#!/usr/bin/env bash

# cli-alpha::deploy starts Kyma installation using the alpha deploy command
#
function cli-alpha::deploy {
  if [[ "$EXECUTION_PROFILE" == "evaluation" ]]; then
	  kyma alpha deploy --ci --profile evaluation --value global.isBEBEnabled=true --verbose
	else
	  kyma alpha deploy --ci --value global.isBEBEnabled=true --verbose
	fi
}